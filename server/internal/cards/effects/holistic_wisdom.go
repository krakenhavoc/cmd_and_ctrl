package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Holistic Wisdom — Enchantment {1}{G}{G} (EDHREC rank 22223):
//
//	"{2}, Exile a card from your hand: Return target card from your
//	 graveyard to your hand if it shares a card type with the card
//	 exiled this way. (Artifact, battle, creature, enchantment,
//	 kindred, instant, land, planeswalker, and sorcery are card
//	 types.)"
//
// The HAND form of #1297's exile cost, and the first CR 602 ability
// whose effect reads the card its cost exiled. The cost is
// AbilityCost.ExileCards from the hand — the same component Cadaverous
// Bloom's mana ability pays (#1283), not a discard, so madness never
// sees the exiled card. The effect reads which card that was off the
// payment record (Context.Exiled, PaidCost.Exiled) and looks it up in
// exile, where the cost put it; by resolution nothing on the board
// could otherwise say which card was paid.
//
// The comparison is on printed card types (CR 205.2a), the only types a
// card in a graveyard or in exile has. The target is re-checked at
// resolution (CR 608.2b), and an exiled card that can no longer be
// found — none can leave exile in the window, but the lookup is the
// honest answer — returns nothing rather than guessing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7e108285-52da-473c-accd-d48e646a49c0",
		Name:         "Holistic Wisdom",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, Exile a card from your hand: Return target card from your graveyard to your hand if it shares a card type with the card exiled this way.",
			Cost:    Plus(ManaCost("{2}"), ExileFromHand(1, "a card", nil)),
			Targets: TargetCardInGraveyard("target card in your graveyard", YouOwn()),
			Effect:  holisticWisdomReturn,
		}},
	})
}

// holisticWisdomReturn is the effect: return the target if it still
// shares a card type with the card the cost exiled.
func holisticWisdomReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	exiled := ctx.Exiled()
	if len(exiled) == 0 {
		return nil
	}
	paid, ok := g.LookupCardForEffect(exiled[0])
	if !ok {
		return nil
	}
	for _, ref := range ctx.LegalTargets() {
		if ref.Kind != game.TargetCard {
			continue
		}
		target, ok := g.LookupCardForEffect(ref.ID)
		if !ok || !sharesACardType(paid, target) {
			return nil
		}
		return ReturnFromGraveyard{Target: ref.ID, Dest: game.ZoneHand}.Apply(ctx)
	}
	return nil
}

// cardTypesCR205 is the card-type list CR 205.2a prints, in its order.
// "tribal" is kindred's pre-2024 name, still on older printings' type
// lines.
var cardTypesCR205 = []string{
	"artifact", "battle", "creature", "enchantment", "kindred", "tribal",
	"instant", "land", "planeswalker", "sorcery",
}

// sharesACardType reports whether `a` and `b` have at least one card
// type in common (CR 205.2a). Subtypes and supertypes do not count.
func sharesACardType(a, b game.Card) bool {
	for _, t := range cardTypesCR205 {
		if a.HasCardType(t) && b.HasCardType(t) {
			return true
		}
	}
	return false
}
