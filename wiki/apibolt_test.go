package wiki

import (
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
)

var (
	bones = Page{Title: "Bones", Links: []string{"Dogs", "Skeletons"}, Linkers: []string{"Dogs", "Doctors"}}
	cats  = Page{Title: "Cats", Links: []string{"Dogs", "Mice"}, Linkers: []string{"Dogs", "Mice", "Humans"}}
	dogs  = Page{Title: "Dogs", Links: []string{"Bones", "Cats"}, Linkers: []string{"Bones", "Cats"}}
	mice  = Page{Title: "Mice", Links: []string{"Cheese", "Cats"}, Linkers: []string{"Cats"}}

	noneBlob = Blob(nil)
	aBlob    = Blob{
		"a_key": []byte("a_val"),
	}
	bBlob = Blob{
		"b_key": []byte("b_val"),
	}
	bothBlob = Blob{
		"a_key": []byte("a_val"),
		"b_key": []byte("b_val"),
	}
	deleteBlobA = Blob{
		"a_key": nil,
	}
	deleteBlobB = Blob{
		"b_key": nil,
	}
	deleteBothBlob = Blob{
		"a_key": nil,
		"b_key": nil,
	}
)

func TestSaveLoadBasic(t *testing.T) {
	runTest := func(expected Page) {
		pr := tempPageRepository()

		err := pr.SavePage(expected)
		if err != nil {
			t.Errorf("Basic SavePage errored with: %s", err)
		}

		actual, err := pr.LoadPage(expected.Title)
		if err != nil {
			t.Errorf("Basic LoadPage errored with: %s", err)
		}

		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("expected: '%#v', actual: '%#v'", expected, actual)
		}
	}

	pages := []Page{
		Page{Title: "Cats", Links: []string{"Dogs", "Mice", ""}, Linkers: []string{"Dogs", "Humans"}},
		Page{Title: "Empty", Links: []string{}},
		Page{Title: "Nil", Links: nil},
	}

	for _, page := range pages {
		runTest(page)
	}
}

func TestSaveLoadOverwrite(t *testing.T) {
	pr := tempPageRepository()
	fullCats := Page{Title: "Cats", Links: []string{"Dogs", "Mice"}}
	emptyCats := Page{Title: "Cats", Links: []string{}}
	fullCatsRedir := Page{Title: "CatsRedir", Redirect: "Cats", Links: []string{"Cats"}}
	emptyCatsRedir := Page{Title: "CatsRedir", Redirect: "Cats", Links: []string{}}

	err := pr.SavePages([]Page{fullCats, fullCatsRedir})
	if err != nil {
		t.Errorf("Basic SavePage errored with: %s", err)
	}
	err = pr.SavePages([]Page{emptyCats, emptyCatsRedir})
	if err != nil {
		t.Errorf("Overwrite SavePage errored with: %s", err)
	}

	actual, err := pr.LoadPage("Cats")
	if err != nil {
		t.Errorf("Overwrite LoadPage errored with: %s", err)
	}
	if !reflect.DeepEqual(emptyCats, actual) {
		t.Errorf("expected: '%#v', actual: '%#v'", emptyCats, actual)
	}

	// use NextPage as a hack to get the direct page value for CatsRedir
	actual, err = pr.NextPage("Cats")
	if err != nil {
		t.Errorf("Overwrite LoadPage errored with: %s", err)
	}
	if !reflect.DeepEqual(emptyCatsRedir, actual) {
		t.Errorf("expected: '%#v', actual: '%#v'", emptyCatsRedir, actual)
	}
}

func TestSaveLoadBlob(t *testing.T) {
	pr := tempPageRepository()

	page := Page{Title: "Cats", Links: []string{"Dogs", "Mice"}}
	expected := page
	mustSavePage(t, pr, page)
	assertLoadedPageIs(t, pr, page.Title, expected)

	page.Blob = aBlob
	expected.Blob = aBlob
	mustSavePage(t, pr, page)
	assertLoadedPageIs(t, pr, page.Title, expected)

	page.Blob = bBlob
	expected.Blob = bothBlob
	mustSavePage(t, pr, page)
	assertLoadedPageIs(t, pr, page.Title, expected)

	page.Blob = deleteBlobA
	expected.Blob = bBlob
	mustSavePage(t, pr, page)
	assertLoadedPageIs(t, pr, page.Title, expected)

	page.Blob = deleteBlobB
	expected.Blob = noneBlob
	mustSavePage(t, pr, page)
	assertLoadedPageIs(t, pr, page.Title, expected)

	page.Blob = bothBlob
	expected.Blob = bothBlob
	mustSavePage(t, pr, page)
	assertLoadedPageIs(t, pr, page.Title, expected)

	page.Blob = noneBlob
	expected.Blob = bothBlob
	mustSavePage(t, pr, page)
	assertLoadedPageIs(t, pr, page.Title, expected)

	page.Blob = deleteBothBlob
	expected.Blob = noneBlob
	mustSavePage(t, pr, page)
	assertLoadedPageIs(t, pr, page.Title, expected)
}

func TestMissingLoad(t *testing.T) {
	pr := tempPageRepository()

	_, err := pr.LoadPage("Missing")

	if err == nil || !strings.Contains(err.Error(), "No entry for title 'Missing'") {
		t.Errorf("Failed to indicate missing page was missing (err: %#v)", err)
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

	err := pr.SavePages([]Page{bones, cats, dogs})
	if err != nil {
		t.Errorf("Bulk SavePages errored with: %s", err)
	}
	err = pr.SavePage(mice)
	if err != nil {
		t.Errorf("SavePage errored with: %s", err)
	}

	loadedBones, err := pr.LoadPage("Bones")
	if err != nil {
		t.Errorf("LoadPage errored with: %s", err)
	}
	if !reflect.DeepEqual(loadedBones, bones) {
		t.Errorf("expected: '%#v', actual: '%#v'", bones, loadedBones)
	}

	loadedCatsDogsMice, err := pr.LoadPages([]string{"Cats", "Dogs", "Mice"})
	if err != nil {
		t.Errorf("Bulk LoadPages errored with: %s", err)
	}
	expected := []Page{cats, dogs, mice}
	if !reflect.DeepEqual(expected, loadedCatsDogsMice) {
		t.Errorf("expected: '%#v', actual: '%#v'", expected, loadedCatsDogsMice)
	}
}

func TestNextPage(t *testing.T) {
	pr := tempPageRepository()

	// don't insert alphabetized
	err := pr.SavePages([]Page{mice, dogs, bones, cats})
	if err != nil {
		t.Errorf("Bulk SavePages errored with: %s", err)
	}

	first, err := pr.FirstPage()
	if err != nil {
		t.Errorf("FirstPage errored with: %s", err)
	}
	if !reflect.DeepEqual(bones, first) {
		t.Errorf("expected: '%#v', actual: '%#v'", bones, first)
	}

	actual, err := pr.NextPage(first.Title)
	if err != nil {
		t.Errorf("FirstPage errored with: %s", err)
	}
	if !reflect.DeepEqual(cats, actual) {
		t.Errorf("expected: '%#v', actual: '%#v'", cats, actual)
	}

	actualPair, err := pr.NextPages(first.Title, 2)
	if err != nil {
		t.Errorf("NextPages errored with: %s", err)
	}
	expected := []Page{cats, dogs}
	if !reflect.DeepEqual(expected, actualPair) {
		t.Errorf("expected: '%#v', actual: '%#v'", expected, actualPair)
	}

	// test running off the end with NextPages
	actualThree, err := pr.NextPages(first.Title, 4)
	if err != nil {
		t.Errorf("NextPages errored with: %s", err)
	}
	expected = []Page{cats, dogs, mice}
	if !reflect.DeepEqual(expected, actualThree) {
		t.Errorf("expected: '%#v', actual: '%#v'", expected, actualThree)
	}

	// test running off the end with NextPage
	actual, err = pr.NextPage(actualThree[len(actualThree)-1].Title)
	if err != nil {
		t.Errorf("NextPage errored with: %s", err)
	}
	if !reflect.DeepEqual(Page{}, actual) {
		t.Errorf("expected: '%#v', actual: '%#v'", Page{}, actual)
	}
}

func TestDeleteTitle(t *testing.T) {
	pr := tempPageRepository()

	err := pr.SavePage(cats)
	if err != nil {
		t.Errorf("SavePage errored with: %s", err)
	}

	err = pr.DeleteTitle(cats.Title)
	if err != nil {
		t.Errorf("Failed to DeleteTitle, errored with: %s", err)
	}

	_, err = pr.LoadPage(cats.Title)
	if err == nil || !strings.Contains(err.Error(), "No entry for title 'Cats'") {
		t.Errorf("Failed to indicate deleted page was missing (err: %#v)", err)
	}
}

func mustSavePage(t *testing.T, pr PageRepository, page Page) {
	err := pr.SavePage(page)
	if err != nil {
		t.Errorf("failed to save page: %#v", page)
	}
}

func assertLoadedPageIs(t *testing.T, pr PageRepository, title string, expected Page) {
	actual, err := pr.LoadPage(title)
	if err != nil {
		t.Errorf("LoadPage errored with: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected: '%#v', actual: '%#v'", expected, actual)
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
