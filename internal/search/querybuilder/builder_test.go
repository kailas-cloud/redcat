package querybuilder

import "testing"

func TestCategoriesOR(t *testing.T) {
	if got := CategoriesOR("category_ids", nil); got != "*" { t.Fatalf("want * got %q", got) }
	if got := CategoriesOR("category_ids", []string{"a","","b"}); got != "@category_ids:{a|b}" { t.Fatalf("got %q", got) }
}

func TestKNN(t *testing.T) {
	f := CategoriesOR("category_ids", []string{"a","b"})
	q := KNN(f, 100, "location")
	want := "@category_ids:{a|b}=>[KNN 100 @location $vec]"
	if q != want { t.Fatalf("want %q got %q", want, q) }
}
