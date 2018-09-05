package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

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

func main() {
	// read feed
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL("https://f5n.org/blog/index.xml")
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
	jsonFile, err := os.Open("./dump.json")
	if err != nil {
		fmt.Println("### Cache file not found.")
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
			// @todo run
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

		jsonOut, err := os.Create("./dump.json")
		if err != nil {
			fmt.Println("### Could not create cache file.")
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
