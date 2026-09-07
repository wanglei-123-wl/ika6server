package catalog

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestHomeStatsUsePublishedGamesOnly(t *testing.T) {
	store := NewStore()
	if _, err := store.AddGame("@draft", "待审核", "不会计入", "Godot", "Puzzle", "MIT"); err != nil {
		t.Fatal(err)
	}
	stats := store.HomeStats()
	if stats["projects"] != 2 {
		t.Fatalf("projects = %v, want 2", stats["projects"])
	}
	if stats["contributors"] != 2 {
		t.Fatalf("contributors = %v, want 2", stats["contributors"])
	}
}

func TestHotPostsRankByInteractionHeat(t *testing.T) {
	store := NewStore()
	item, err := store.AddPost("builder", "热帖", "交流", "测试", nil, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.ReviewPost(item.ID, "approved"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.LikePost(item.ID, 1); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 6000; i++ {
		store.ViewPost(item.ID)
	}
	hot := store.HotPosts()
	if len(hot) == 0 || hot[0].ID != item.ID {
		t.Fatalf("hot post id = %#v, want %d first", hot, item.ID)
	}
}

func TestConcurrentLikeGameCountsUserOnce(t *testing.T) {
	store := NewStore()
	before := store.games[0].Likes

	var changed atomic.Int64
	runConcurrent(t, 50, func() {
		_, ok, err := store.LikeGame(1, 42)
		if err != nil {
			t.Errorf("LikeGame failed: %v", err)
			return
		}
		if ok {
			changed.Add(1)
		}
	})

	game, ok := store.Game(1)
	if !ok {
		t.Fatal("game not found")
	}
	if changed.Load() != 1 {
		t.Fatalf("changed likes = %d, want 1", changed.Load())
	}
	if game.Likes != before+1 {
		t.Fatalf("game likes = %d, want %d", game.Likes, before+1)
	}
}

func TestConcurrentLikePostCountsDifferentUsers(t *testing.T) {
	store := NewStore()
	post, err := store.AddPost("builder", "并发点赞", "交流", "测试", nil, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.ReviewPost(post.ID, "approved"); err != nil {
		t.Fatal(err)
	}

	const users = 40
	runConcurrentIndexed(t, users, func(i int) {
		if _, _, err := store.LikePost(post.ID, int64(i+1)); err != nil {
			t.Errorf("LikePost failed: %v", err)
		}
	})

	got, ok := store.Post(post.ID)
	if !ok {
		t.Fatal("post not found")
	}
	if got.Likes != users {
		t.Fatalf("post likes = %d, want %d", got.Likes, users)
	}
}

func TestConcurrentLikeReplyCountsUserOnce(t *testing.T) {
	store := NewStore()
	reply, err := store.AddReply("tester", "回复点赞测试", 1)
	if err != nil {
		t.Fatal(err)
	}

	var changed atomic.Int64
	runConcurrent(t, 50, func() {
		_, ok, err := store.LikeReply(reply.ID, 7)
		if err != nil {
			t.Errorf("LikeReply failed: %v", err)
			return
		}
		if ok {
			changed.Add(1)
		}
	})

	replies, ok := store.Replies(1)
	if !ok || len(replies) != 1 {
		t.Fatalf("replies = %#v, found = %v", replies, ok)
	}
	if changed.Load() != 1 {
		t.Fatalf("changed likes = %d, want 1", changed.Load())
	}
	if replies[0].Likes != 1 {
		t.Fatalf("reply likes = %d, want 1", replies[0].Likes)
	}
}

func TestConcurrentViewPostDoesNotLoseIncrements(t *testing.T) {
	store := NewStore()
	post, err := store.AddPost("builder", "并发浏览", "交流", "测试", nil, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.ReviewPost(post.ID, "approved"); err != nil {
		t.Fatal(err)
	}

	const views = 250
	runConcurrent(t, views, func() {
		if _, ok := store.ViewPost(post.ID); !ok {
			t.Error("ViewPost did not find post")
		}
	})

	got, ok := store.Post(post.ID)
	if !ok {
		t.Fatal("post not found")
	}
	if parseCount(got.Views) != views {
		t.Fatalf("post views = %q, want %d", got.Views, views)
	}
}

func TestConcurrentDownloadRepoDoesNotLoseIncrements(t *testing.T) {
	store := NewStore()
	store.repos = []Repo{{ID: 99, Name: "fixture", Downloads: "0"}}

	const downloads = 250
	runConcurrent(t, downloads, func() {
		if _, err := store.DownloadRepo(99); err != nil {
			t.Errorf("DownloadRepo failed: %v", err)
		}
	})

	repo, ok := store.Repo(99)
	if !ok {
		t.Fatal("repo not found")
	}
	if parseCount(repo.Downloads) != downloads {
		t.Fatalf("repo downloads = %q, want %d", repo.Downloads, downloads)
	}
}

func runConcurrent(t *testing.T, workers int, work func()) {
	t.Helper()
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			work()
		}()
	}
	wg.Wait()
}

func runConcurrentIndexed(t *testing.T, workers int, work func(int)) {
	t.Helper()
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		i := i
		go func() {
			defer wg.Done()
			work(i)
		}()
	}
	wg.Wait()
}
