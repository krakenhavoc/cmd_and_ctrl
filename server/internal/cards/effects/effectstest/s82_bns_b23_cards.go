package effectstest

import (
	"testing"
)

const (
	BnsB23GuardianIdol   = "829d2e11-192f-4093-99f8-e13767b321a2"
	BnsB23Puzzleknot     = "c19d2f54-2c3a-4cbe-8328-061d4b51f574"
	BnsB23Memento        = "887de1fc-8fbd-472a-bcf2-4f36ea9376d1"
	BnsB23Solvent        = "f34ba913-cf9a-48fe-b5fe-68bf504693b2"
	BnsB23Whirlermaker   = "bdf28770-0962-4317-822f-c187bc6706d1"
	BnsB23WitchingWell   = "baadfc17-42fd-4771-aabe-027ef7b3bf69"
	BnsB23AncestralBlade = "ab28a801-6ab1-4db2-8ae2-c24bb4c10ebb"
	BnsB23Kitesail       = "68fb8f21-4403-4829-8826-f2418664197e"
	BnsB23GorgonFlail    = "91dab3e2-039e-4db9-94e3-1146bced18ed"
	BnsB23LeoninScimitar = "cde26d69-f3e7-4dd0-a53b-cd0ec812d717"
	BnsB23RelicAxe       = "d851312f-8b4e-4d8e-bfee-c470853be930"
)

func BnsB23Pump(t *testing.T, oracle, name string, wantP, wantT int, keyword string) {
	t.Helper()
	g := NewCatalogGame(t)
	me := g.Seats[0]
	bear := PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	eq := PushCatalogPermanent(g, me.ID, name, "Artifact — Equipment", oracle, false)
	if p, tt := EffectivePTOf(t, g, bear); p != 2 || tt != 2 {
		t.Fatalf("unequipped bear is %d/%d", p, tt)
	}
	AdvanceToMain(t, g)
	FloatMana(t, g, me, "{C}{C}")
	EquipTo(t, g, me.ID, eq, bear)
	if p, tt := EffectivePTOf(t, g, bear); p != 2+wantP || tt != 2+wantT {
		t.Errorf("%s: equipped bear is %d/%d, want %d/%d", name, p, tt, 2+wantP, 2+wantT)
	}
	if keyword != "" && !HasKeywordOf(g, bear, keyword) {
		t.Errorf("%s: no %s on the equipped creature", name, keyword)
	}
}
