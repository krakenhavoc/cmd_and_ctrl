package game

import (
	"testing"

	"github.com/google/uuid"
)

// protection_doesnt_remove_test.go — #4843, CR 702.16n: an instance of
// protection that "doesn't remove" one Aura, or all Auras. The exemption
// rides on the token, so it is per instance, and only the CR 704.5m
// attachment check reads it.

func TestProtectionSparingAuraTokenRoundTrips(t *testing.T) {
	aura := uuid.New()
	tok := ProtectionSparingAura(ProtectionFromColor("W"), aura)
	q, ok := ParseProtectionQuality(tok)
	if !ok {
		t.Fatalf("%q does not parse", tok)
	}
	if q.Kind != ProtectionQualityColor || q.Value != "W" || q.Printed != "white" {
		t.Errorf("parsed %+v, want protection from white printed as \"white\"", q)
	}
	if q.SparesAura != aura || q.SparesAuras {
		t.Errorf("exemption = %s / %v, want %s only", q.SparesAura, q.SparesAuras, aura)
	}
	if q.Token() != tok {
		t.Errorf("Token() = %q, want %q", q.Token(), tok)
	}

	all := ProtectionSparingAuras(ProtectionFromColor("U"))
	q, ok = ParseProtectionQuality(all)
	if !ok || !q.SparesAuras || q.SparesAura != uuid.Nil || q.Value != "U" || q.Token() != all {
		t.Errorf("%q parsed as %+v (%v)", all, q, ok)
	}
}

func TestProtectionSparingRefusesWhatItCannotMint(t *testing.T) {
	if got := ProtectionSparingAura("protection from monocolored creatures", uuid.New()); got != "" {
		t.Errorf("an unparseable quality minted %q", got)
	}
	if got := ProtectionSparingAura(ProtectionFromColor("W"), uuid.Nil); got != "" {
		t.Errorf("a nil Aura minted %q", got)
	}
	if got := ProtectionSparingAuras(ProtectionSparingAuras(ProtectionFromColor("W"))); got != "" {
		t.Errorf("a doubled exemption minted %q", got)
	}
	for _, bad := range []string{
		"protection from white (doesn't remove nothing)",
		"protection from white (doesn't remove " + uuid.New().String(),
		"protection from white (doesn't remove " + uuid.Nil.String() + ")",
	} {
		if _, ok := ParseProtectionQuality(bad); ok {
			t.Errorf("%q parsed; a malformed exemption must mint no protection", bad)
		}
	}
}

// CR 702.16n: the sparing instance leaves its own Aura alone and nothing
// else; a second instance of the same quality removes the Aura as normal.
func TestProtectedFromAttachmentHonoursTheExemptionPerInstance(t *testing.T) {
	ward := &Card{InstanceID: uuid.New(), TypeLine: "Enchantment — Aura", Colors: []string{"W"}}
	other := &Card{InstanceID: uuid.New(), TypeLine: "Enchantment — Aura", Colors: []string{"W"}}
	equip := &Card{InstanceID: uuid.New(), TypeLine: "Artifact — Equipment", Colors: []string{"W"}}
	sparing := ProtectionSparingAura(ProtectionFromColor("W"), ward.InstanceID)

	host := func(abilities ...string) *Card {
		return &Card{InstanceID: uuid.New(), TypeLine: "Creature — Bear", Keywords: abilities}
	}

	if ProtectedFromAttachment(host(sparing), ward) {
		t.Error("the sparing instance removed its own Aura")
	}
	if !ProtectedFromAttachment(host(sparing), other) {
		t.Error("the sparing instance spared a different white Aura")
	}
	if !ProtectedFromAttachment(host(sparing), equip) {
		t.Error("the sparing instance spared a white Equipment, which is not an Aura")
	}
	if !ProtectedFromAttachment(host(sparing, ProtectionFromColor("W")), ward) {
		t.Error("a second instance of protection from white did not remove the Aura")
	}
	if !ProtectedFromAttachment(host(sparing, ProtectionSparingAura(ProtectionFromColor("W"), other.InstanceID)), ward) {
		t.Error("another Aura's sparing instance spared this one")
	}
	all := ProtectionSparingAuras(ProtectionFromColor("W"))
	if ProtectedFromAttachment(host(all), ward) || ProtectedFromAttachment(host(all), other) {
		t.Error("\"doesn't remove Auras\" removed an Aura")
	}
	if !ProtectedFromAttachment(host(all), equip) {
		t.Error("\"doesn't remove Auras\" spared an Equipment")
	}

	// Every other consumer still reads the sparing instance as plain
	// protection from white (CR 702.16b, e, f).
	if !ProtectedFrom(host(sparing), SourceCharacteristics(ward)) {
		t.Error("the sparing instance does not protect from a white source")
	}
}

// CR 702.16p, Benevolent Blessing: "This effect doesn't remove Auras and
// Equipment you control that are already attached to it." The instance
// spares an Aura or Equipment its named player controls, and nothing else.
func TestProtectionSparingAttachedOfSparesOnlyThatPlayersAurasAndEquipment(t *testing.T) {
	me, them := uuid.New(), uuid.New()
	mineAura := &Card{InstanceID: uuid.New(), TypeLine: "Enchantment — Aura", Colors: []string{"W"}, Controller: me}
	mineEquip := &Card{InstanceID: uuid.New(), TypeLine: "Artifact — Equipment", Colors: []string{"W"}, Controller: me}
	theirsAura := &Card{InstanceID: uuid.New(), TypeLine: "Enchantment — Aura", Colors: []string{"W"}, Controller: them}
	theirsEquip := &Card{InstanceID: uuid.New(), TypeLine: "Artifact — Equipment", Colors: []string{"W"}, Controller: them}
	tok := ProtectionSparingAttachedOf(ProtectionFromColor("W"), me)

	q, ok := ParseProtectionQuality(tok)
	if !ok || q.SparesAttachedOf != me || q.SparesAura != uuid.Nil || q.SparesAuras || q.Value != "W" || q.Token() != tok {
		t.Fatalf("%q parsed as %+v (%v)", tok, q, ok)
	}
	host := &Card{InstanceID: uuid.New(), TypeLine: "Creature — Bear", Keywords: []string{tok}}
	if ProtectedFromAttachment(host, mineAura) || ProtectedFromAttachment(host, mineEquip) {
		t.Error("the exemption removed an Aura or Equipment its player controls")
	}
	if !ProtectedFromAttachment(host, theirsAura) || !ProtectedFromAttachment(host, theirsEquip) {
		t.Error("the exemption spared another player's Aura or Equipment")
	}
	twice := &Card{InstanceID: uuid.New(), TypeLine: "Creature — Bear", Keywords: []string{tok, ProtectionFromColor("W")}}
	if !ProtectedFromAttachment(twice, mineAura) {
		t.Error("a second instance of protection from white did not remove the Aura")
	}
	if !ProtectedFrom(host, SourceCharacteristics(mineAura)) {
		t.Error("the sparing instance does not protect from a white source")
	}
	if got := ProtectionSparingAttachedOf(ProtectionFromColor("W"), uuid.Nil); got != "" {
		t.Errorf("a nil player minted %q", got)
	}
	if got := ProtectionSparingAttachedOf(tok, me); got != "" {
		t.Errorf("a doubled exemption minted %q", got)
	}
	for _, bad := range []string{
		"protection from white (doesn't remove Auras and Equipment controlled by nobody)",
		"protection from white (doesn't remove Auras and Equipment controlled by " + uuid.Nil.String() + ")",
	} {
		if _, ok := ParseProtectionQuality(bad); ok {
			t.Errorf("%q parsed; a malformed exemption must mint no protection", bad)
		}
	}
}
