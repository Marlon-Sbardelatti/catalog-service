package mongodb

import "testing"

func TestGenerateURI(t *testing.T) {
	client := &Client{
		Host:     "localhost",
		Port:     "27017",
		Username: "mongoadmin",
		Password: "secret",
		Database: "catalog-service",
	}

	want := "mongodb://mongoadmin:secret@localhost:27017/?authSource=admin"
	if got := client.GenerateURI(); got != want {
		t.Errorf("GenerateURI() = %q, want %q", got, want)
	}
}
