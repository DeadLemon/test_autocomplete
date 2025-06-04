package main

import (
	"bufio"
	"container/list"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
)

// WordFreq holds a word and its frequency.
type WordFreq struct {
	Word      string
	Frequency int
}

// parseWordFreq parses a line "<word> <frequency>".
func parseWordFreq(line string) (WordFreq, error) {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return WordFreq{}, fmt.Errorf("invalid word frequency line")
	}
	var wf WordFreq
	wf.Word = parts[0]
	_, err := fmt.Sscanf(parts[1], "%d", &wf.Frequency)
	return wf, err
}

// loadWordFreq loads word frequencies from file.
func loadWordFreq(path string) ([]WordFreq, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var res []WordFreq
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		wf, err := parseWordFreq(scanner.Text())
		if err != nil {
			continue
		}
		res = append(res, wf)
	}
	return res, scanner.Err()
}

// LRUCache is a simple LRU cache for prefix suggestions.
type LRUCache struct {
	cap   int
	mu    sync.Mutex
	list  *list.List
	items map[string]*list.Element
}

type cacheEntry struct {
	key   string
	value []WordFreq
}

func newLRUCache(cap int) *LRUCache {
	return &LRUCache{
		cap:   cap,
		list:  list.New(),
		items: make(map[string]*list.Element),
	}
}

func (c *LRUCache) Get(key string) ([]WordFreq, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.items[key]; ok {
		c.list.MoveToFront(elem)
		return elem.Value.(*cacheEntry).value, true
	}
	return nil, false
}

func (c *LRUCache) Add(key string, value []WordFreq) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.items[key]; ok {
		elem.Value.(*cacheEntry).value = value
		c.list.MoveToFront(elem)
		return
	}
	elem := c.list.PushFront(&cacheEntry{key: key, value: value})
	c.items[key] = elem
	if c.list.Len() > c.cap {
		old := c.list.Back()
		if old != nil {
			c.list.Remove(old)
			delete(c.items, old.Value.(*cacheEntry).key)
		}
	}
}

// Server is autocomplete tcp server.
type Server struct {
	wordFreq []WordFreq
	cache    *LRUCache
}

func NewServer(wordFreq []WordFreq) *Server {
	return &Server{wordFreq: wordFreq, cache: newLRUCache(128)}
}

func parseCommand(line string) (string, string, error) {
	fields := strings.Fields(strings.TrimSpace(line))
	if len(fields) != 2 || strings.ToLower(fields[0]) != "get" {
		return "", "", fmt.Errorf("Command should match pattern 'get <prefix:str>'")
	}
	prefix := fields[1]
	if len(prefix) < 1 || len(prefix) > 15 {
		return "", "", fmt.Errorf("Command should match pattern 'get <prefix:str>'")
	}
	if !isAlpha(prefix) {
		return "", "", fmt.Errorf("Command should match pattern 'get <prefix:str>'")
	}
	return fields[0], prefix, nil
}

func isAlpha(s string) bool {
	for _, r := range s {
		if r < 'A' || (r > 'Z' && r < 'a') || r > 'z' {
			return false
		}
	}
	return true
}

func (s *Server) getSuggestions(prefix string) []WordFreq {
	if val, ok := s.cache.Get(prefix); ok {
		return val
	}
	var res []WordFreq
	for _, wf := range s.wordFreq {
		if strings.HasPrefix(wf.Word, prefix) {
			res = append(res, wf)
		}
	}
	sort.Slice(res, func(i, j int) bool {
		if res[i].Frequency == res[j].Frequency {
			return res[i].Word < res[j].Word
		}
		return res[i].Frequency > res[j].Frequency
	})
	if len(res) > 10 {
		res = res[:10]
	}
	s.cache.Add(prefix, res)
	return res
}

func (s *Server) handleConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	fmt.Fprintln(conn, "This is autocomplete service.")
	fmt.Fprintln(conn, "Service accepts commands, which")
	fmt.Fprintln(conn, "matches 'get <prefix>' pattern")
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		_, prefix, err := parseCommand(line)
		if err != nil {
			fmt.Fprintln(conn, err.Error())
			continue
		}
		suggestions := s.getSuggestions(prefix)
		if len(suggestions) == 0 {
			fmt.Fprintln(conn, "Suggestions not found")
			continue
		}
		for _, item := range suggestions {
			fmt.Fprintf(conn, "-> %s\n", item.Word)
		}
	}
}

func main() {
	filename := flag.String("filename", "word_freq.txt", "path to word freq file")
	port := flag.Int("port", 10000, "listen port")
	flag.Parse()

	words, err := loadWordFreq(*filename)
	if err != nil {
		log.Fatalf("failed to load word freq file: %v", err)
	}

	srv := NewServer(words)

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("Server running on %s", ln.Addr())

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept error: %v", err)
			continue
		}
		go srv.handleConn(context.Background(), conn)
	}
}
