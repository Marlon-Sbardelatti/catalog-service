package product

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/textproto"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type fakeRepository struct {
	created  *Product
	found    *Product
	previous *Product
	all      []Product
	set      bson.M
	err      error
}

func (f *fakeRepository) Create(_ context.Context, p *Product) error {
	f.created = p
	return f.err
}
func (f *fakeRepository) FindByID(context.Context, bson.ObjectID) (*Product, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.found == nil {
		return nil, ErrProductNotFound
	}
	return f.found, nil
}
func (f *fakeRepository) FindAll(context.Context) ([]Product, error) { return f.all, nil }
func (f *fakeRepository) Update(_ context.Context, _ bson.ObjectID, set bson.M) (*Product, error) {
	f.set = set
	if f.err != nil {
		return nil, f.err
	}
	return f.previous, nil
}
func (f *fakeRepository) Delete(context.Context, bson.ObjectID) error {
	return f.err
}

type fakeStorage struct {
	uploads   int
	deletes   []string
	err       error
	deleteErr error
}

func (f *fakeStorage) Upload(context.Context, string, io.Reader, int64, string) error {
	f.uploads++
	return f.err
}

func (f *fakeStorage) Delete(_ context.Context, key string) error {
	f.deletes = append(f.deletes, key)
	return f.deleteErr
}

func (f *fakeStorage) PublicURL(key string) string {
	return "https://cdn.test/" + key
}

func imageFile(t *testing.T, contentType string) *multipart.FileHeader {
	t.Helper()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="image"; filename="pixel.png"`)
	header.Set("Content-Type", contentType)

	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatalf("creating part: %v", err)
	}
	if _, err := part.Write([]byte("fake-bytes")); err != nil {
		t.Fatalf("writing part: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("closing writer: %v", err)
	}

	form, err := multipart.NewReader(body, writer.Boundary()).ReadForm(1 << 20)
	if err != nil {
		t.Fatalf("reading form: %v", err)
	}
	t.Cleanup(func() { _ = form.RemoveAll() })

	return form.File["image"][0]
}

func TestCreate(t *testing.T) {
	validReq := CreateProductRequest{Name: "Camisa", Price: 299}

	tests := []struct {
		name          string
		req           CreateProductRequest
		contentType   string
		uploadErr     error
		wantErr       error
		wantUploads   int
		wantPersisted bool
	}{
		{
			name:          "creates product and uploads image",
			req:           validReq,
			contentType:   "image/png",
			wantUploads:   1,
			wantPersisted: true,
		},
		{
			name:        "rejects invalid product before uploading",
			req:         CreateProductRequest{Name: "  ", Price: 299},
			contentType: "image/png",
			wantErr:     ErrInvalidProduct,
			wantUploads: 0,
		},
		{
			name:        "rejects non-image upload",
			req:         validReq,
			contentType: "application/pdf",
			wantErr:     ErrInvalidProduct,
			wantUploads: 0,
		},
		{
			name:          "does not persist when upload fails",
			req:           validReq,
			contentType:   "image/png",
			uploadErr:     errors.New("s3 down"),
			wantUploads:   1,
			wantPersisted: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := &fakeRepository{}
			storage := &fakeStorage{err: tt.uploadErr}
			service := NewProductService(repository, storage)

			product, err := service.Create(context.Background(), tt.req, imageFile(t, tt.contentType))

			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("Create() error = %v, want %v", err, tt.wantErr)
			}
			if tt.uploadErr != nil && err == nil {
				t.Fatal("Create() error = nil, want upload error")
			}
			if storage.uploads != tt.wantUploads {
				t.Errorf("uploads = %d, want %d", storage.uploads, tt.wantUploads)
			}
			if (repository.created != nil) != tt.wantPersisted {
				t.Errorf("persisted = %v, want %v", repository.created != nil, tt.wantPersisted)
			}

			if tt.wantPersisted {
				if !strings.HasPrefix(product.ImageKey, "products/") {
					t.Errorf("ImageKey = %q, want products/ prefix", product.ImageKey)
				}
				if want := "https://cdn.test/" + product.ImageKey; product.ImageURL != want {
					t.Errorf("ImageURL = %q, want %q", product.ImageURL, want)
				}
			}
		})
	}
}

func TestFindAllSetsImageURL(t *testing.T) {
	repository := &fakeRepository{all: []Product{{ImageKey: "products/a"}}}
	service := NewProductService(repository, &fakeStorage{})

	products, err := service.FindAll(context.Background())
	if err != nil {
		t.Fatalf("FindAll() error = %v", err)
	}

	if want := "https://cdn.test/products/a"; products[0].ImageURL != want {
		t.Errorf("ImageURL = %q, want %q", products[0].ImageURL, want)
	}
}

func TestUpdate(t *testing.T) {
	t.Run("applies fields and deletes the previous image", func(t *testing.T) {
		previous := &Product{ID: bson.NewObjectID(), Name: "Old", Price: 100, ImageKey: "products/old"}
		repository := &fakeRepository{previous: previous}
		storage := &fakeStorage{}
		service := NewProductService(repository, storage)

		product, err := service.Update(context.Background(), previous.ID,
			UpdateProductRequest{Name: "  Camisa  ", Price: 299}, imageFile(t, "image/png"))
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		if product.Name != "Camisa" || product.Price != 299 {
			t.Errorf("product = %+v, want updated name and price", product)
		}
		if !strings.HasPrefix(product.ImageKey, "products/") || product.ImageKey == previous.ImageKey {
			t.Errorf("ImageKey = %q, want a new key", product.ImageKey)
		}
		if want := "https://cdn.test/" + product.ImageKey; product.ImageURL != want {
			t.Errorf("ImageURL = %q, want %q", product.ImageURL, want)
		}
		if len(storage.deletes) != 1 || storage.deletes[0] != previous.ImageKey {
			t.Errorf("deletes = %v, want [%s]", storage.deletes, previous.ImageKey)
		}
		if _, ok := repository.set["updated_at"]; !ok {
			t.Error("set missing updated_at")
		}
	})

	t.Run("keeps the previous image when no file is sent", func(t *testing.T) {
		previous := &Product{ID: bson.NewObjectID(), Name: "Old", Price: 100, ImageKey: "products/old"}
		repository := &fakeRepository{previous: previous}
		storage := &fakeStorage{}
		service := NewProductService(repository, storage)

		product, err := service.Update(context.Background(), previous.ID, UpdateProductRequest{Price: 150}, nil)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		if product.Price != 150 || product.ImageKey != previous.ImageKey || product.Name != "Old" {
			t.Errorf("product = %+v, want only price changed", product)
		}
		if _, ok := repository.set["image_key"]; ok {
			t.Error("set must not contain image_key")
		}
		if len(storage.deletes) != 0 {
			t.Errorf("deletes = %v, want none", storage.deletes)
		}
	})

	t.Run("rejects blank name before uploading", func(t *testing.T) {
		storage := &fakeStorage{}
		service := NewProductService(&fakeRepository{}, storage)

		_, err := service.Update(context.Background(), bson.NewObjectID(),
			UpdateProductRequest{Name: "   "}, imageFile(t, "image/png"))
		if !errors.Is(err, ErrInvalidProduct) {
			t.Fatalf("Update() error = %v, want ErrInvalidProduct", err)
		}
		if storage.uploads != 0 {
			t.Errorf("uploads = %d, want 0", storage.uploads)
		}
	})

	t.Run("rejects an empty update", func(t *testing.T) {
		service := NewProductService(&fakeRepository{}, &fakeStorage{})

		_, err := service.Update(context.Background(), bson.NewObjectID(), UpdateProductRequest{}, nil)
		if !errors.Is(err, ErrInvalidProduct) {
			t.Fatalf("Update() error = %v, want ErrInvalidProduct", err)
		}
	})

	t.Run("deletes the fresh upload when the update fails", func(t *testing.T) {
		repository := &fakeRepository{err: ErrProductNotFound}
		storage := &fakeStorage{}
		service := NewProductService(repository, storage)

		_, err := service.Update(context.Background(), bson.NewObjectID(),
			UpdateProductRequest{Price: 10}, imageFile(t, "image/png"))
		if !errors.Is(err, ErrProductNotFound) {
			t.Fatalf("Update() error = %v, want ErrProductNotFound", err)
		}

		key, _ := repository.set["image_key"].(string)
		if len(storage.deletes) != 1 || storage.deletes[0] != key {
			t.Errorf("deletes = %v, want [%s]", storage.deletes, key)
		}
	})

	t.Run("does not fail when deleting the previous image fails", func(t *testing.T) {
		previous := &Product{ID: bson.NewObjectID(), ImageKey: "products/old"}
		storage := &fakeStorage{deleteErr: errors.New("s3 down")}
		service := NewProductService(&fakeRepository{previous: previous}, storage)

		if _, err := service.Update(context.Background(), previous.ID,
			UpdateProductRequest{Price: 10}, imageFile(t, "image/png")); err != nil {
			t.Fatalf("Update() error = %v, want nil", err)
		}
	})
}

func TestErrorsAreNotSwallowed(t *testing.T) {
	boom := errors.New("mongo is down")

	service := NewProductService(&fakeRepository{err: boom}, &fakeStorage{})

	if _, err := service.FindByID(context.Background(), bson.NewObjectID()); !errors.Is(err, boom) {
		t.Errorf("FindByID() error = %v, want %v", err, boom)
	}
	if err := service.Delete(context.Background(), bson.NewObjectID()); !errors.Is(err, boom) {
		t.Errorf("Delete() error = %v, want %v", err, boom)
	}
}
