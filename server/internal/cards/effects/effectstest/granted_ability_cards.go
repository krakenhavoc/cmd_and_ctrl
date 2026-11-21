package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// granted_ability_cards_test.go — the first cards on ADR 0093's seam
// (the ADR's PR 2): static and attached grants of mana and activated
// abilities. Each test asserts behaviour — who has the ability, who
// may activate it, what it does — rather than a label.

const (
	GaCryptolithRiteOracle   = "043f869d-b11c-4c0d-9591-2bf0df7bde55"
	GaChromaticLanternOracle = "539f5396-d99a-417d-a84c-dff7930b5900"
	GaGemhideSliverOracle    = "2c09ca09-8e62-4fe3-9b3d-61573dd2ffbc"
	GaManaweftSliverOracle   = "bd47398d-da35-4a09-8754-771af91b14f4"
	GaNecroticSliverOracle   = "9655569d-bfa5-4665-9371-9f275b8d223e"
	GaRishkarOracle          = "761021ce-4559-464e-aa03-85c2fe78e267"
	GaJaheiraOracle          = "0aeeb0d7-15a3-4722-ad67-8fc7fe89220d"
	GaInsidiousRootsOracle   = "d75b2b8e-05c9-47da-b359-a867256d78ea"
	GaGreatDivideGuideOracle = "79e69a91-d580-47fb-be76-1e32c50d2fa0"
	GaTheWorldTreeOracle     = "3437d504-bf62-4c27-b15f-f6330182ff7e"
	GaParadiseMantleOracle   = "c1121b83-1ba2-473d-89c9-e3bbd4529072"
	GaSquirrelNestOracle     = "2d584333-b25e-4291-a2c9-80c6d9f8732a"
)

// gaPut seeds a permanent that has been under its controller's control
// since the turn began (so a {T} is not sick).
func GaPut(g *game.Game, owner uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: owner, Controller: owner,
	})
}

// layeredCard is the live battlefield card with its layer cache fresh.
func GaLayered(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var out game.Card
	var ok bool
	g.ReadSnapshot(func() { out, ok = g.LookupCardForEffect(id) })
	if !ok {
		t.Fatalf("%s is not in the game", id)
	}
	return out
}

// grantedManaRowsOf counts a permanent's GRANTED mana ability rows.
func GrantedManaRowsOf(g *game.Game, id uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			return
		}
		_, origins := game.ManaAbilitiesWithOrigins(c)
		for _, o := range origins {
			if o.Granted() {
				n++
			}
		}
	})
	return n
}

// grantedActivatedIndex is the index of a permanent's first granted
// activated ability, or -1.
func GrantedActivatedIndex(t *testing.T, g *game.Game, id uuid.UUID) (int, string) {
	t.Helper()
	_, origins := game.ActivatedAbilitiesWithOrigins(GaLayered(t, g, id))
	for i, o := range origins {
		if o.Granted() {
			return i, o.Ref
		}
	}
	return -1, ""
}

// answerTriggerTargets answers a multi-target trigger's pick_target
// prompt with the given cards.
func AnswerTriggerTargets(t *testing.T, g *game.Game, chooser uuid.UUID, cards ...uuid.UUID) {
	t.Helper()
	pick := LatestPickTarget(g, chooser)
	if pick == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	refs := make([]game.TargetRef, 0, len(cards))
	for _, c := range cards {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: c})
	}
	if err := g.ResolvePickTargets(pick.ID, chooser, refs); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
}

// tapGrantedMana taps a permanent for its first granted mana ability,
// answering the colour pick with `color` when one is asked.
func TapGrantedMana(t *testing.T, g *game.Game, player, card uuid.UUID, color string) error {
	t.Helper()
	_, origins := game.ManaAbilitiesWithOrigins(GaLayered(t, g, card))
	for i, o := range origins {
		if !o.Granted() {
			continue
		}
		if err := g.ActivateManaAbility(player, card, i, game.ManaAbilityParams{Ref: o.Ref}); err != nil {
			return err
		}
		for _, ch := range g.PendingChoices {
			if ch != nil && ch.Kind == game.PendingChoiceMana && ch.Chooser == player {
				if err := g.ResolveManaChoice(ch.ID, player, color); err != nil {
					t.Fatalf("ResolveManaChoice: %v", err)
				}
			}
		}
		return nil
	}
	t.Fatalf("%s has no granted mana ability", card)
	return nil
}
