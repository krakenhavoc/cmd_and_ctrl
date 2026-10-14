package shoplinks

import (
	"strings"
	"testing"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestFromEnvUnsetIsNoShops(t *testing.T) {
	shops, err := FromEnv(env(nil))
	if err != nil {
		t.Fatalf("FromEnv: %v", err)
	}
	if shops == nil || len(shops) != 0 {
		t.Fatalf("shops = %#v, want an empty, non-nil list", shops)
	}
}

func TestFromEnvReadsEachRetailerInOrder(t *testing.T) {
	shops, err := FromEnv(env(map[string]string{
		"CMDCTRL_SHOP_CARDKINGDOM": " {url}&partner=cmdctrl ",
		"CMDCTRL_SHOP_TCGPLAYER":   "https://partner.example/c/123?u={url}",
	}))
	if err != nil {
		t.Fatalf("FromEnv: %v", err)
	}
	want := []Shop{
		{ID: "tcgplayer", Name: "TCGplayer", Template: "https://partner.example/c/123?u={url}"},
		{ID: "cardkingdom", Name: "Card Kingdom", Template: "{url}&partner=cmdctrl"},
	}
	if len(shops) != len(want) {
		t.Fatalf("shops = %#v, want %#v", shops, want)
	}
	for i := range want {
		if shops[i] != want[i] {
			t.Errorf("shops[%d] = %#v, want %#v", i, shops[i], want[i])
		}
	}
}

func TestFromEnvNamesTheBadVariable(t *testing.T) {
	_, err := FromEnv(env(map[string]string{
		"CMDCTRL_SHOP_TCGPLAYER":   "https://partner.example/c/123?u={url}",
		"CMDCTRL_SHOP_CARDKINGDOM": "http://partner.example/?u={url}",
	}))
	if err == nil || !strings.Contains(err.Error(), "CMDCTRL_SHOP_CARDKINGDOM") {
		t.Fatalf("err = %v, want one naming CMDCTRL_SHOP_CARDKINGDOM", err)
	}
	if strings.Contains(err.Error(), "CMDCTRL_SHOP_TCGPLAYER") {
		t.Errorf("err = %v names the good variable too", err)
	}
}

func TestValidateTemplate(t *testing.T) {
	good := []string{
		"{url}",
		"{url}&partner=cmdctrl",
		"{url}&partner=cmdctrl&utm_source=cmdctrl",
		"https://partner.example/c/123?u={url}",
		"https://partner.example/c/123/{url}",
		"https://partner.example/c/123?subId1=c2%2Cbuy&u={url}",
	}
	for _, tpl := range good {
		if err := ValidateTemplate(tpl); err != nil {
			t.Errorf("ValidateTemplate(%q) = %v, want ok", tpl, err)
		}
	}
	bad := map[string]string{
		"no placeholder":           "https://partner.example/c/123",
		"two placeholders":         "https://partner.example/?u={url}&v={url}",
		"plain http":               "http://partner.example/?u={url}",
		"javascript":               "javascript:alert(1)//{url}",
		"no scheme":                "partner.example/?u={url}",
		"no host":                  "https:///?u={url}",
		"userinfo":                 "https://user@partner.example/?u={url}",
		"placeholder in host":      "https://{url}.example/",
		"placeholder as host":      "https://{url}",
		"space":                    "https://partner.example/?u={url}&x=a b",
		"quote":                    `https://partner.example/?u={url}"`,
		"angle bracket":            "https://partner.example/<x>?u={url}",
		"stray brace":              "https://partner.example/?u={url}&x={y",
		"non-ascii":                "https://partner.example/?u={url}&x=é",
		"direct with no pairs":     "{url}&",
		"direct with a path":       "{url}/more",
		"direct with nameless":     "{url}&=x",
		"direct with bad escape":   "{url}&x=%zz",
		"direct with semicolon":    "{url}&a;b=1",
		"text before direct start": "x{url}",
		"too long":                 "https://partner.example/?u={url}&x=" + strings.Repeat("a", MaxTemplateBytes),
	}
	for name, tpl := range bad {
		if err := ValidateTemplate(tpl); err == nil {
			t.Errorf("%s: ValidateTemplate(%q) = nil, want an error", name, tpl)
		}
	}
}
