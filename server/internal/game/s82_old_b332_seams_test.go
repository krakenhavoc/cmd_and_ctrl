package game

import "testing"

// Seam built for S82 OLD batch 332 (#5551): ManaRestrictTurnFaceUpMorphOrManifest
// (Qarsi Deceiver).

// The mana pays the morph route and a manifested creature's mana cost, and
// nothing else: not a disguise cost, not a cloaked creature's mana cost, not a
// cast, not an activation.
func TestManaRestrictTurnFaceUpMorphOrManifestAdmitsOnlyThoseRoutes(t *testing.T) {
	tags := []string{ManaRestrictTurnFaceUpMorphOrManifest}
	for _, tc := range []struct {
		name  string
		route TurnFaceUpRoute
		kind  FaceDownKind
		want  bool
	}{
		{"morph", TurnFaceUpByMorph, FaceDownMorphed, true},
		{"morph cost of a manifested card", TurnFaceUpByMorph, FaceDownManifested, true},
		{"mana cost of a manifested creature", TurnFaceUpByManaCost, FaceDownManifested, true},
		{"mana cost of a cloaked creature", TurnFaceUpByManaCost, FaceDownCloaked, false},
		{"disguise", TurnFaceUpByDisguise, FaceDownDisguised, false},
		{"disguise of a cloaked card", TurnFaceUpByDisguise, FaceDownCloaked, false},
		{"no route named", "", FaceDownMorphed, false},
	} {
		ctx := ManaSpendContext{Purpose: SpendPurposeTurnFaceUp, TurnFaceUpRoute: tc.route, TurnFaceUpKind: tc.kind}
		if got := ctx.allows(tags); got != tc.want {
			t.Errorf("%s: allows = %v, want %v", tc.name, got, tc.want)
		}
	}
	if ManaSpendForCast(Card{ManaCost: "{1}{R}", TypeLine: "Creature — Bear"}).allows(tags) {
		t.Error("admitted a cast")
	}
	if ManaSpendForAbility(Card{TypeLine: "Artifact"}).allows(tags) {
		t.Error("admitted an activation")
	}
}

// specialActionSpendContextFor names the route and the face-down kind for a turn
// face up, and nothing for any other kind.
func TestSpecialActionSpendContextForNamesTheRoute(t *testing.T) {
	card := Card{FaceDownKind: FaceDownManifested}
	ctx := specialActionSpendContextFor(SpecialAction{Kind: SpecialActionTurnFaceUp, Route: TurnFaceUpByManaCost}, card)
	if ctx.Purpose != SpendPurposeTurnFaceUp || ctx.TurnFaceUpRoute != TurnFaceUpByManaCost || ctx.TurnFaceUpKind != FaceDownManifested {
		t.Errorf("turn face up context = %+v", ctx)
	}
	if got := specialActionSpendContextFor(SpecialAction{Kind: SpecialActionForetell, Route: TurnFaceUpByMorph}, card); got.TurnFaceUpRoute != "" || got.TurnFaceUpKind != "" {
		t.Errorf("foretell context carries a turn-face-up route: %+v", got)
	}
}
