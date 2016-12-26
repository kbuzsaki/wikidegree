package wiki

import "testing"

func TestTitlePath(t *testing.T) {
	var path TitlePath

	firstHead := path.Head()
	if firstHead != "" {
		t.Errorf("expected head %#v, was %#v", "", firstHead)
	}

	titles := []string{"dog", "bat", "zoo"}
	for _, title := range titles {
		path = path.Catted(title)

		head := path.Head()
		if head != title {
			t.Errorf("expected head %#v, was %#v", title, head)
		}
	}

	lastHead := path.Head()
	lastTitle := titles[len(titles)-1]
	if lastHead != lastTitle {
		t.Errorf("expected lastHead %#v, was %#v", lastTitle, lastHead)
	}
}

func TestNormalizeTitle(t *testing.T) {
	testPairs := [][2]string{
		{"", ""},
		{"dog", "Dog"},
		{"dog cat", "Dog_cat"},
		{"dog Cat", "Dog_Cat"},
		{"DOG", "DOG"},
		{"dOG", "DOG"},
		{"dOG cAT", "DOG_cAT"},
		{"DOG CAT", "DOG_CAT"},
		{"dog cat bat ball", "Dog_cat_bat_ball"},
	}

	for _, testPair := range testPairs {
		input := testPair[0]
		expected := testPair[1]

		actual := NormalizeTitle(input)
		if expected != actual {
			t.Errorf("expected normalized title %#v, was %#v", expected, actual)
		}
	}
}
