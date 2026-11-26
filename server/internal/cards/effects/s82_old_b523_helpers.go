package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b523_helpers.go — Eventide, activated abilities (S82 older sets,
// batch 523, #5590).

// S82OldB523SpellTargetsYou is "spell that targets you": the spell on the stack
// has the asking player among its announced targets. It reads the stack item
// for the card, so a spell that has already left the stack is not one.
func S82OldB523SpellTargetsYou() CardPredicate {
	return func(g *game.Game, caster uuid.UUID, c game.Card) bool {
		item := g.StackItemForEffect(c.InstanceID)
		if item == nil {
			return false
		}
		for _, t := range item.Targets {
			if t.Kind == game.TargetPlayer && t.ID == caster {
				return true
			}
		}
		return false
	}
}

// S82OldB523CreaturesOnTheBattlefield counts every creature on the battlefield,
// whoever controls it (Altar Golem). It reads the cards' types directly so a
// layer pass can call it.
func S82OldB523CreaturesOnTheBattlefield(g *game.Game) int {
	n := 0
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].IsCreature() {
			n++
		}
	}
	return n
}

// S82OldB523BlueInAdditionToItsOtherColors is "Target permanent becomes blue in
// addition to its other colors until end of turn" (Indigo Faerie): a layer 5
// addition, not a replacement of its colours (CR 613.1e).
func S82OldB523BlueInAdditionToItsOtherColors(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return ScopedEffectFor{
			Target:   t.ID,
			Mods:     []game.Mod{game.AddColorsMod(ctx.TextColors("U")...)},
			Duration: DurationUntilEndOfTurn(ctx),
			Label:    "Indigo Faerie — becomes blue in addition to its other colors until end of turn",
		}.Apply(ctx)
	}
	return nil
}

// S82OldB523PumpByTappedPower is "This creature gets +X/+0 until end of turn,
// where X is the power of the creature tapped this way" (Impelled Giant). The
// tapped creature's power is read as the ability resolves, or as it last
// existed if it has left (CR 608.2h); a power of zero or less adds nothing.
func S82OldB523PumpByTappedPower(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	power, ok := ctx.TappedPower()
	if !ok || power <= 0 || !OnBattlefield(g, ctx.Source()) {
		return nil
	}
	return BoostUntilEOT{Target: ctx.Source(), Power: power, Label: "Impelled Giant — +X/+0 for the tapped creature's power"}.Apply(ctx)
}

// S82OldB523TapFiveUntappedCreatures is the cost "Tap five untapped creatures
// you control" (Altar Golem). Without "another", the source may be one of them.
func S82OldB523TapFiveUntappedCreatures() game.AbilityCost {
	return game.AbilityCost{TapOthers: &game.TapOthersCost{
		Count:  5,
		Filter: TargetPermanent("five untapped creatures you control", Creature()),
		Label:  "five untapped creatures you control",
	}}
}
