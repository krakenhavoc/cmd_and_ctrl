package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b022_helpers.go — Secret Lair Drop, activated abilities (S82 older
// sets, batch 022, #5462): the shared shapes of Charix, the Raging Isle,
// Sakashima's Student, Leshrac's Sigil and Voldaren Pariah.

// S82OldB022AlsoANinja is the `except` clause of "except it's a Ninja in
// addition to its other creature types" (CR 707.9b): Sakashima's Student.
func S82OldB022AlsoANinja(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
	v.AddSubtype("Ninja")
}

// S82OldB022PlusXMinusXForIslands is Charix's "{3}: Charix gets +X/-X until
// end of turn, where X is the number of Islands you control": X is counted as
// the ability resolves (CR 608.2h) and fixed from then on.
func S82OldB022PlusXMinusXForIslands(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	island := ctx.TextSubtype("Island") // a word of the source's text (CR 612.2)
	x := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.HasSubtype(island) {
			x++
		}
	}
	if x == 0 {
		return nil
	}
	return BoostUntilEOT{Target: item.SourceCardID, Power: x, Toughness: -x, Label: "Charix, the Raging Isle — +X/-X"}.Apply(ctx)
}

// S82OldB022TargetOpponentSacrificesThree is Abolisher of Bloodlines' trigger
// body: "target opponent sacrifices three creatures of their choice". The one
// player named by the target chooses (CR 701.21a); with fewer than three
// creatures they sacrifice what they have. The choice is not a target, so
// hexproof is irrelevant to the creatures.
func S82OldB022TargetOpponentSacrificesThree(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := uuid.Nil
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			player = t.ID
			break
		}
	}
	if player == uuid.Nil || !g.CanEffectCauseSacrificeForEffect(player) {
		return nil
	}
	candidates := g.EffectSacrificeCandidatesForEffect(player, SacrificeSpec("a creature", Creature()))
	if len(candidates) == 0 {
		return nil
	}
	k := min(3, len(candidates))
	return g.ChoosePermanentsInOrderSacrificeForEffect(item.SourceCardID, []game.ChooseCardsPrompt{{
		Chooser: player, FromPlayer: player, Source: item.SourceCardID,
		Question: "Abolisher of Bloodlines — sacrifice three creatures",
		Cards:    candidates, Min: k, Max: k, Zone: game.ZoneBattlefield,
		Destination: game.ChooseSacrifice,
	}})
}

// S82OldB022LeshracsSigilPaid is the paid half of Leshrac's Sigil's trigger:
// "look at that player's hand and choose a card from it. The player discards
// that card". The player is the opponent whose green spell triggered it.
var S82OldB022LeshracsSigilPaid = PayUnlessBranches("pay-unless/s82-old-b022-leshracs-sigil",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		caster := TriggeringActor(ctx.Item)
		if caster == uuid.Nil {
			return nil
		}
		return ChooseFromRevealedHand{Player: caster, Look: true, Label: "card"}.Apply(ctx)
	}, nil)

// S82OldB022LeshracsSigilOffer is the trigger's resolution: its controller may
// pay {B}{B}.
func S82OldB022LeshracsSigilOffer(g *game.Game, item *game.StackItem) error {
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{B}{B}",
		Question: "Leshrac's Sigil — pay {B}{B} to look at that player's hand and make them discard a card you choose?",
		Then:     S82OldB022LeshracsSigilPaid,
	}.Apply(NewContext(g, item))
}
