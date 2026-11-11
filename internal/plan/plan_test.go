package plan

import "testing"

func TestLibraryLoads(t *testing.T) {
	lib, err := LoadLibrary()
	if err != nil {
		t.Fatal(err)
	}
	if len(lib.Modules) != 8 {
		t.Fatalf("modules = %d, want 8", len(lib.Modules))
	}
	for _, m := range lib.Modules {
		if m.ID == "" || m.Title == "" || m.Summary == "" {
			t.Fatalf("module %q has empty fields", m.ID)
		}
		if len(m.Sessions) != 2 {
			t.Fatalf("module %s has %d sessions, want 2", m.ID, len(m.Sessions))
		}
	}
}

func TestComposeShape(t *testing.T) {
	lib, err := LoadLibrary()
	if err != nil {
		t.Fatal(err)
	}
	p := lib.Compose(lib.ParseSelection("m1,m3,m5,m7"))
	if len(p.Items) != 15 {
		t.Fatalf("items = %d, want 15", len(p.Items))
	}
	if p.Items[0].Kind != "workshop" {
		t.Fatalf("item 0 kind = %s", p.Items[0].Kind)
	}
	if p.Items[len(p.Items)-1].Kind != "demo" {
		t.Fatalf("last item kind = %s", p.Items[len(p.Items)-1].Kind)
	}
	for i, it := range p.Items {
		if it.N != i {
			t.Fatalf("item %d has N=%d", i, it.N)
		}
		if it.Title == "" || it.Goal == "" || it.Artifact == "" {
			t.Fatalf("item %d has an empty field", i)
		}
	}
	// only the eight module blocks vary with the selection
	mod := 0
	for _, it := range p.Items {
		if it.Kind == "module" {
			mod++
		}
	}
	if mod != 8 {
		t.Fatalf("module blocks = %d, want 8", mod)
	}
}

func TestParseSelectionPadsAndDrops(t *testing.T) {
	lib, err := LoadLibrary()
	if err != nil {
		t.Fatal(err)
	}
	got := lib.ParseSelection("m2,bogus,m9")
	if len(got) != 4 {
		t.Fatalf("selection = %v (len %d), want 4", got, len(got))
	}
	if got[0] != "m2" {
		t.Fatalf("selection[0] = %s, want m2", got[0])
	}
	for _, id := range got {
		if _, ok := lib.Get(id); !ok {
			t.Fatalf("selection contains unknown id %q", id)
		}
	}
}
