package quiz

import "testing"

func validModules() map[string]bool {
	return map[string]bool{"m1": true, "m2": true, "m3": true, "m4": true, "m5": true, "m6": true, "m7": true, "m8": true}
}

func TestLoadHasTenQuestions(t *testing.T) {
	d, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Questions) != 10 {
		t.Fatalf("questions = %d, want 10", len(d.Questions))
	}
	for _, q := range d.Questions {
		if len(q.Options) < 2 {
			t.Fatalf("question %d has %d options", q.ID, len(q.Options))
		}
	}
}

func TestScoreSelectsFour(t *testing.T) {
	d, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	answers := make([]string, len(d.Questions))
	for i := range answers {
		answers[i] = "A"
	}
	res, err := Score(answers, d.Questions, validModules())
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Selected) != 4 {
		t.Fatalf("selected = %v (len %d), want 4", res.Selected, len(res.Selected))
	}
	for _, id := range res.Selected {
		if !validModules()[id] {
			t.Fatalf("selected unknown module %q", id)
		}
	}
}

func TestScoreIsDeterministic(t *testing.T) {
	d, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	answers := make([]string, len(d.Questions))
	for i := range answers {
		answers[i] = "B"
	}
	a, _ := Score(answers, d.Questions, validModules())
	b, _ := Score(answers, d.Questions, validModules())
	if quizJoin(a.Selected) != quizJoin(b.Selected) {
		t.Fatalf("not deterministic: %v vs %v", a.Selected, b.Selected)
	}
}

func quizJoin(s []string) string {
	out := ""
	for i, v := range s {
		if i > 0 {
			out += ","
		}
		out += v
	}
	return out
}

func TestScoreRejectsBadInput(t *testing.T) {
	d, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Score([]string{"A"}, d.Questions, validModules()); err == nil {
		t.Fatal("expected an error for a short answer set")
	}
	answers := make([]string, len(d.Questions))
	for i := range answers {
		answers[i] = "A"
	}
	answers[0] = "Z"
	if _, err := Score(answers, d.Questions, validModules()); err == nil {
		t.Fatal("expected an error for an unknown option key")
	}
}