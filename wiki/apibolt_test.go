package wiki

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestSaveLoadBasic(t *testing.T) {
	pr := tempPageRepository()
	expected := Page{Title: "Cats", Links: []string{"Dogs", "Mice"}}

	err := pr.SavePage(expected)
	if err != nil {
		t.Errorf("Basic SavePage errored with: %s", err)
	}

	actual, err := pr.LoadPage("Cats")
	if err != nil {
		t.Errorf("Basic LoadPage errored with: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected: '%#v', actual: '%#v'", expected, actual)
	}
}

func TestMissingLoad(t *testing.T) {
	pr := tempPageRepository()

	_, err := pr.LoadPage("Missing")

	if err == nil || err.Error() != "No entry for title 'Missing'" {
		t.Errorf("Failed to error when loading while closed")
	}
}

func TestClosedLoad(t *testing.T) {
	pr := tempPageRepository()

	pr.Close()
	_, err := pr.LoadPage("Foo")

	if err == nil || err.Error() != "Connection closed" {
		t.Errorf("Failed to error when loading while closed")
	}
}

func TestSaveLoadRedirect(t *testing.T) {
	pr := tempPageRepository()
	redirect := Page{Title: "CatsRedir", Redirect: "Cats"}
	cats := Page{Title: "Cats", Links: []string{"Dogs", "Mice"}}

	err := pr.SavePages([]Page{cats, redirect})
	if err != nil {
		t.Errorf("Redirect SavePages errored with: %s", err)
	}

	expected := Page{Redirector: "CatsRedir", Title: cats.Title, Links: cats.Links}
	actual, err := pr.LoadPage("CatsRedir")
	if err != nil {
		t.Errorf("Redirect LoadPage errored with: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected: '%#v', actual: '%#v'", expected, actual)
	}
}

func TestBulkSaveLoad(t *testing.T) {
	pr := tempPageRepository()
	bones := Page{Title: "Bones", Links: []string{"Dogs", "Skeletons"}}
	cats := Page{Title: "Cats", Links: []string{"Dogs", "Mice"}}
	dogs := Page{Title: "Dogs", Links: []string{"Bones", "Cats"}}
	mice := Page{Title: "Mice", Links: []string{"Cheese", "Cats"}}

	err := pr.SavePages([]Page{bones, cats, dogs})
	if err != nil {
		t.Errorf("Redirect SavePages errored with: %s", err)
	}
	err = pr.SavePage(mice)
	if err != nil {
		t.Errorf("Redirect SavePage errored with: %s", err)
	}

	loadedBones, err := pr.LoadPage("Bones")
	if err != nil {
		t.Errorf("Redirect LoadPage errored with: %s", err)
	}
	if !reflect.DeepEqual(loadedBones, bones) {
		t.Errorf("expected: '%#v', actual: '%#v'", bones, loadedBones)
	}

	loadedCatsDogsMice, err := pr.LoadPages([]string{"Cats", "Dogs", "Mice"})
	if err != nil {
		t.Errorf("Redirect LoadPages errored with: %s", err)
	}
	expected := []Page{cats, dogs, mice}
	if !reflect.DeepEqual(expected, loadedCatsDogsMice) {
		t.Errorf("expected: '%#v', actual: '%#v'", expected, loadedCatsDogsMice)
	}
}

func tempPageRepository() PageRepository {
	pr, err := GetBoltPageRepository(tempfile())
	if err != nil {
		panic(err)
	}
	return pr
}

// gets a temporary file to use for the db
// copied from https://github.com/boltdb/bolt/blob/a5aec31dc3d13cbd7c0e6faca7489835b0b7e27a/db_test.go#L1628
func tempfile() string {
	f, err := ioutil.TempFile("", "wikidegree-")
	if err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
	if err := os.Remove(f.Name()); err != nil {
		panic(err)
	}
	return f.Name()
}
