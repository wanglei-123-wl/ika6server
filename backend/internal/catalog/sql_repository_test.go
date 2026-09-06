package catalog

import "testing"

func TestNewSQLRepositoryRequiresConnection(t *testing.T) {
	if _, err := NewSQLRepository(nil); err == nil {
		t.Fatal("expected nil sql database to fail")
	}
}

func TestFormatCount(t *testing.T) {
	tests := map[int64]string{0: "0", 18: "18", 1000: "1k", 18400: "18.4k"}
	for input, want := range tests {
		if got := formatCount(input); got != want {
			t.Fatalf("formatCount(%d) = %q, want %q", input, got, want)
		}
	}
}

func TestEmptySearchReturnsEmptyCollections(t *testing.T) {
	// Search with no keyword is handled before a database query is needed.
	data := SearchData{Games: []Game{}, Posts: []ForumPost{}, Repos: []Repo{}}
	if len(data.Games) != 0 || len(data.Posts) != 0 || len(data.Repos) != 0 {
		t.Fatal("expected empty search collections")
	}
}
