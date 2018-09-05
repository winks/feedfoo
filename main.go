package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"syscall"

	"github.com/mmcdole/gofeed"
)

type Posts struct {
	Posts []Post `json:"posts"`
}

type Post struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Updated string `json:"updated"`
}

var feedUrl string
var jsonCacheFile string
var printHelp bool
var tail []string

func showHelp() {
	fmt.Println("usage: feedfoo [OPTS] -- madonctl --toot %%TEXT%%")
	flag.PrintDefaults()
	os.Exit(0)
}

func initArgs() {
	flag.BoolVar(&printHelp, "help", false, "print help and exit")
	flag.StringVar(&feedUrl, "feed", "", "the feed to check")
	flag.StringVar(&jsonCacheFile, "cache", "", "the json cache file")

	flag.Parse()

	if printHelp || feedUrl == "" || jsonCacheFile == "" {
		showHelp()
	}

	// this is the command to run
	tail = flag.Args()
	if len(tail) < 1 {
		os.Exit(3)
	}
}

func main() {
	initArgs()
	// get the real path of the executable
	// exec.Command() would do LookPath, but this is for error handling
	binary, lookErr := exec.LookPath(tail[0])
	// this is not syscall.Exec(), so we need to shift the first element off
	tail = tail[1:]
	if lookErr != nil {
		fmt.Println(lookErr)
		os.Exit(1)
	}

	// read feed
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedUrl)
	if err != nil {
		fmt.Println("Error1")
		return
	}
	fmt.Println("# Items for: ", feed.Title)
	fmt.Println("# ---")

	// stuff
	var cachedPosts Posts
	var cachedLookup map[string]int
	cachedLookup = make(map[string]int)
	var newPosts Posts

	// read JSON
	jsonFile, err := os.Open(jsonCacheFile)
	if err != nil {
		fmt.Printf("### Cache file %s not found.\n", jsonCacheFile)
	} else {
		defer jsonFile.Close()
		bytes, err := ioutil.ReadAll(jsonFile)
		if err != nil {
			fmt.Println("### Error3")
			return
		}
		json.Unmarshal(bytes, &cachedPosts)
	}

	// parse JSON
	var post Post
	fmt.Println("# Cached posts:")
	for i := 0; i < len(cachedPosts.Posts); i++ {
		post = cachedPosts.Posts[i]
		line := "" + post.Updated + " :: " + post.Link + " | " + post.Title
		fmt.Println(line)
		cachedLookup[post.Link] = i
	}
	fmt.Println("# --- total cached: ", len(cachedLookup))

	var item gofeed.Item
	for i := range feed.Items {
		item = *feed.Items[i]
		line := "" + (*item.UpdatedParsed).String() + " :: " + item.Link + " | " + item.Title
		fmt.Println(line)
		if _, ok := cachedLookup[item.Link]; ok {
			fmt.Println("# Cached: ", line)
		} else {
			fmt.Println("# New   : ", line)
			text := "New blog post: " + item.Title + " " + item.Link
			retval := run(text, binary, tail)
			if retval != 0 {
				log.Fatal("Error executing, not updating cache file.")
				return
			}
			var newPost Post
			newPost.Link = item.Link
			newPost.Title = item.Title
			newPost.Updated = item.Updated
			newPosts.Posts = append(newPosts.Posts, newPost)
		}
		fmt.Println("# -")

		if i >= 4 {
			break
		}
	}
	fmt.Println("# total new: ", len(newPosts.Posts))

	// save to JSON
	if len(newPosts.Posts) > 0 {
		for i := 0; i < len(newPosts.Posts); i++ {
			cachedPosts.Posts = append(cachedPosts.Posts, newPosts.Posts[i])
		}

		jsonOut, err := os.Create(jsonCacheFile)
		if err != nil {
			fmt.Printf("### Could not create cache file %s.\n", jsonCacheFile)
			return
		}
		defer jsonOut.Close()
		fmt.Println("# new total: ", len(cachedPosts.Posts))
		jsonData, err := json.Marshal(cachedPosts)
		if err != nil {
			fmt.Println("### Could not create cache file.2")
			return
		}
		jsonOut.Write(jsonData)
		jsonOut.Close()
	}
}

func run(text string, binary string, tail []string) int {
	reg, err := regexp.Compile("%%TEXT%%")
	if err != nil {
		fmt.Println(err)
		return 2
	}

	for k, v := range tail {
		tail[k] = reg.ReplaceAllString(v, text)
	}

	fmt.Printf("#> %s %s\n", binary, tail)
	cmd := exec.Command(binary, tail...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err != nil {
		log.Fatal(err)
		return 2
	}
	execErr := cmd.Start()
	if err != nil {
		log.Fatal(err)
		return 2
	}
	execErr = cmd.Wait()

	// the exit code was zero
	if execErr != nil {
		exitCode2 := 0
		if exitError, ok := execErr.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode2 = ws.ExitStatus()
		} else {
			exitCode2 = 2
		}
		return exitCode2
	}
	// exit code was 0
	return 0
}
