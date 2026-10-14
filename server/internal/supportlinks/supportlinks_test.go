package supportlinks

import (
	"strings"
	"testing"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestFromEnvUnsetIsNoLinks(t *testing.T) {
	l, err := FromEnv(env(nil))
	if err != nil || l != (Links{}) {
		t.Fatalf("FromEnv = %+v, %v; want empty, nil", l, err)
	}
	l, err = FromEnv(env(map[string]string{KofiEnv: "  ", PatreonEnv: ""}))
	if err != nil || l != (Links{}) {
		t.Fatalf("blank FromEnv = %+v, %v; want empty, nil", l, err)
	}
}

func TestFromEnvReadsBoth(t *testing.T) {
	l, err := FromEnv(env(map[string]string{
		KofiEnv:    " https://ko-fi.com/cmdctrl ",
		PatreonEnv: "https://www.patreon.com/c/cmdctrl?x=1",
	}))
	if err != nil {
		t.Fatalf("FromEnv: %v", err)
	}
	want := Links{Kofi: "https://ko-fi.com/cmdctrl", Patreon: "https://www.patreon.com/c/cmdctrl?x=1"}
	if l != want {
		t.Fatalf("links = %+v, want %+v", l, want)
	}
}

func TestFromEnvNamesEveryBadVariable(t *testing.T) {
	_, err := FromEnv(env(map[string]string{
		KofiEnv:    "http://ko-fi.com/x",
		PatreonEnv: "https://patreon.com.evil.example/x",
	}))
	if err == nil {
		t.Fatal("want an error")
	}
	for _, name := range []string{KofiEnv, PatreonEnv} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("err = %v, want it to name %s", err, name)
		}
	}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		raw, domain string
		ok          bool
	}{
		{"https://ko-fi.com/x", "ko-fi.com", true},
		{"https://www.ko-fi.com/x", "ko-fi.com", true},
		{"https://KO-FI.com/x", "ko-fi.com", true},
		{"https://patreon.com/x", "patreon.com", true},
		{"https://www.patreon.com/c/x", "patreon.com", true},
		{"http://ko-fi.com/x", "ko-fi.com", false},
		{"//ko-fi.com/x", "ko-fi.com", false},
		{"ko-fi.com/x", "ko-fi.com", false},
		{"https://ko-fi.com.evil.example/x", "ko-fi.com", false},
		{"https://evilko-fi.com/x", "ko-fi.com", false},
		{"https://app.ko-fi.com/x", "ko-fi.com", false},
		{"https://ko-fi.com@evil.example/x", "ko-fi.com", false},
		{"https://user:pw@ko-fi.com/x", "ko-fi.com", false},
		{"https://ko-fi.com:8443/x", "ko-fi.com", false},
		{"https://ko-fi.com/x y", "ko-fi.com", false},
		{"https://ko-fi.com/<x>", "ko-fi.com", false},
		{"https://patreon.com/x", "ko-fi.com", false},
		{"javascript:alert(1)", "ko-fi.com", false},
		{"https://ko-fi.com/" + strings.Repeat("a", MaxBytes), "ko-fi.com", false},
	}
	for _, c := range cases {
		err := Validate(c.raw, c.domain)
		if (err == nil) != c.ok {
			t.Errorf("Validate(%q, %q) = %v, want ok=%v", c.raw, c.domain, err, c.ok)
		}
	}
}
