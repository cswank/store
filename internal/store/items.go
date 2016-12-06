package store

import (
	"archive/zip"
	"fmt"
	"image/jpeg"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/cswank/store/internal/shopify"
	"github.com/nfnt/resize"
)

var (
	lock  sync.Mutex
	items *Items
)

func DeleteAllProducts() error {
	return db.DeleteAll([]byte("products"))
}

type Items struct {
	home       string
	items      map[string]map[string][]string
	categories []string
}

func NewItems(h string) *Items {
	return &Items{home: h}
}

func GetCategories() []string {
	lock.Lock()
	defer lock.Unlock()
	return items.categories
}

func GetCategory(name string) map[string][]string {
	lock.Lock()
	defer lock.Unlock()
	return items.items[name]
}

func GetCategoryList(name string) []string {
	lock.Lock()
	defer lock.Unlock()
	m := items.items[name]
	l := make([]string, len(m))

	var i int
	for k, _ := range m {
		l[i] = k
		i++
	}
	return l
}

func SetItems(i *Items) {
	lock.Lock()
	items = i
	lock.Unlock()
}

func GetProductID(key string) (int, error) {
	var id int64
	return int(id), db.Get([]byte(key), []byte("products"), func(val []byte) error {
		s := string(val)
		var err error
		id, err = strconv.ParseInt(s, 10, 64)
		return err
	})
}

func (i *Items) Load(src string) error {
	i.items = map[string]map[string][]string{}
	i.categories = []string{}
	return filepath.Walk(src, func(pth string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		trimmed := strings.TrimPrefix(strings.TrimPrefix(pth, src), string(filepath.Separator))
		parts := strings.Split(trimmed, string(filepath.Separator))
		if len(parts) == 0 || strings.HasPrefix(parts[len(parts)-1], ".") {
			return nil
		}

		var cat, subcat, item string
		for j, part := range parts {
			if part == "" {
				return nil
			}
			switch j {
			case 0:
				cat = part
				i.addCategory(part)
			case 1:
				subcat = part
				i.addSubcategory(cat, part)
			case 2:
				item = part
				if err := i.addItem(cat, subcat, part); err != nil {
					return err
				}
			case 3:
				if strings.Contains(pth, "image.jpg") || strings.Contains(pth, "thumb.jpg") {
					return nil
				}

				if err := i.addImage(cat, subcat, item, pth); err != nil {
					return err
				}
			}
		}
		return nil
	})

}

//use shopify api to add product
func (i *Items) addProduct(cat, item string) error {
	err := db.Get([]byte(item), []byte("products"), func(val []byte) error {
		return nil
	})

	if err != nil && err != ErrNotFound {
		return err
	}

	if err == nil { //already exists
		return nil
	}

	id, err := shopify.CreateProduct(item, cat)
	if err != nil {
		return err
	}

	val := []byte(strconv.FormatInt(int64(id), 10))
	return db.Put([]byte(item), val, []byte("products"))
}

func (i *Items) addCategory(p string) {
	_, ok := i.items[p]
	if ok {
		return
	}

	i.categories = append(i.categories, p)
	i.items[p] = map[string][]string{}
}

func (i *Items) addSubcategory(cat, p string) {
	_, ok := i.items[cat][p]
	if ok {
		return
	}
	i.items[cat][p] = []string{}
}

func (i *Items) addItem(cat, subcat, p string) error {
	idx := i.getIndex(cat, subcat, p)

	var err error
	if idx == -1 {
		s := i.items[cat][subcat]
		s = append(s, p)
		i.items[cat][subcat] = s
		err = i.addProduct(cat, p)
	}
	return err
}

func (i *Items) addImage(cat, subcat, item, p string) error {
	f, err := os.Open(p)
	if err != nil {
		return err
	}

	img, err := jpeg.Decode(f)
	if err != nil {
		return err
	}
	f.Close()

	dir, _ := filepath.Split(p)
	full := filepath.Join(dir, "image.jpg")
	thumb := filepath.Join(dir, "thumb.jpg")

	sizes := map[uint]string{
		340: full,
		200: thumb,
	}

	for k, v := range sizes {
		if fileExists(v) {
			continue
		}
		m := resize.Resize(k, 0, img, resize.Lanczos3)
		out, err := os.Create(v)
		if err != nil {
			return err
		}

		jpeg.Encode(out, m, nil)

		out.Close()

		// write new image to file
		jpeg.Encode(out, m, nil)
	}

	return nil

}

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func (i *Items) getIndex(cat, subcat, item string) int {
	for j, it := range i.items[cat][subcat] {
		if it == item {
			return j
		}
	}
	return -1
}

/*
1: get data
2: write to local dir $STORE_HOME/archive/items-<number>.zip
3: unzip to tmpdir
3: parse data (items.Load)
4: if succeeds, delete old $STORE_HOME/items
5: move upzipped tmpdir to $STORE_HOME/items
6: set in memory items to new one
*/

func ImportItems(r io.Reader) error {
	name, err := archiveItems(r)
	if err != nil {
		return err
	}

	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}

	if err := unzip(name, tmp); err != nil {
		return err
	}

	pth := filepath.Join(tmp, "items")
	i := Items{}
	if err := i.Load(pth); err != nil {
		return err
	}

	dst := filepath.Join(cfg.DataDir, "items")
	if err := os.RemoveAll(dst); err != nil {
		return err
	}

	if err := os.Rename(pth, dst); err != nil {
		return err
	}

	SetItems(&i)
	return nil
}

func archiveItems(r io.Reader) (string, error) {
	dir := filepath.Join(cfg.DataDir, "archive")
	if !fileExists(dir) {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return "", err
		}
	}
	name := getArchiveName(dir)
	f, err := os.Create(name)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(f, r)
	f.Close()
	return name, err
}

func getArchiveName(dir string) string {
	files, _ := ioutil.ReadDir(dir)
	return filepath.Join(dir, fmt.Sprintf("items-%d.zip", len(files)))
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}
