package cardshu

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ice Cauldron — Artifact {4} (Ice Age, S82 OLD batch 845, #5636):
//
//	"{X}, {T}: You may exile a nonland card from your hand. You may cast that
//	 card for as long as it remains exiled. Put a charge counter on this
//	 artifact and note the type and amount of mana spent to pay this
//	 activation cost. Activate only if there are no charge counters on this
//	 artifact.
//	 {T}, Remove a charge counter from this artifact: Add this artifact's last
//	 noted type and amount of mana. Spend this mana only to cast the last card
//	 exiled with this artifact."
//
// As the first ability resolves it puts the counter on, notes the mana the
// activation spent (the colour of each unit, read off the paid-cost record and
// kept as a note on the permanent, replacing the last one: Card.Notes under
// game.SetPrefixedNoteForEffect) and forgets the card it last exiled, then
// asks its controller which nonland card in their hand, if any, to exile. The
// exile is ExileWithPermission, so the card may be cast for as long as it stays
// there, and the card's instance is noted as the last one exiled with the
// Cauldron. The second ability is a mana ability (CR 605.1a) whose cost removes
// the counter and which adds exactly the noted mana, restricted to the cast of
// that one card (game.ManaRestrictSpellCard); with no card exiled the mana can
// be spent on nothing. Mana paid on paper, with no recorded colour, is not
// noted (the weaker direction).
//
// No simplification.
func init() {
	const manaNote, cardNote = "cauldron-mana:", "cauldron-card:"
	effects.Register(effects.Spec{
		OracleID:       "2b7e4e70-53b1-4d4f-bdcc-56c2e15f367a",
		Name:           "Ice Cauldron",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsOwnActivations,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}, {T}: You may exile a nonland card from your hand. You may cast that card for as long as it remains exiled. Put a charge counter on this artifact and note the type and amount of mana spent to pay this activation cost. Activate only if there are no charge counters on this artifact.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{X}"), effects.TapCost()),
			Condition: func(g *game.Game, _, source uuid.UUID) bool {
				c, ok := g.LookupCardForEffect(source)
				return ok && c.Counters[game.CounterCharge] <= 0
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if effects.SourceIsNewObject(g, item) || !effects.OnBattlefield(g, item.SourceCardID) {
					return nil
				}
				var spent strings.Builder
				for _, sym := range []string{"W", "U", "B", "R", "G", "C"} {
					for i := item.Paid.Spent().Count(sym); i > 0; i-- {
						spent.WriteString("{" + sym + "}")
					}
				}
				g.SetPrefixedNoteForEffect(item.SourceCardID, manaNote, spent.String())
				g.SetPrefixedNoteForEffect(item.SourceCardID, cardNote, "")
				if err := g.AddCounterForEffect(item.SourceCardID, game.CounterCharge, 1); err != nil {
					return err
				}
				p := g.PlayerByIDForEffect(item.Controller)
				if p == nil || p.Hand == nil {
					return nil
				}
				var cards []uuid.UUID
				for _, c := range p.Hand.Cards {
					if !c.IsLand() {
						cards = append(cards, c.InstanceID)
					}
				}
				if len(cards) == 0 {
					return nil
				}
				g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
					Chooser:  item.Controller,
					Source:   item.SourceCardID,
					Question: "Ice Cauldron — you may exile a nonland card from your hand",
					Cards:    cards,
					Min:      0,
					Max:      1,
					LiveZone: game.ZoneHand,
					ThenKey:  iceCauldronExile,
					Asked:    game.NewAskedItem(item),
				})
				return nil
			},
		}},
		ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{
				Tap:            true,
				RemoveCounters: effects.RemoveCountersFromThis(game.CounterCharge, 1).RemoveCounters,
			},
			ProducedFunc: func(g *game.Game, _, source uuid.UUID) string {
				v, _ := g.PrefixedNoteForEffect(source, manaNote)
				return v
			},
			RestrictionsFunc: func(g *game.Game, _, source uuid.UUID) []string {
				id := uuid.Nil
				if v, ok := g.PrefixedNoteForEffect(source, cardNote); ok {
					if parsed, err := uuid.Parse(v); err == nil {
						id = parsed
					}
				}
				return []string{game.ManaRestrictSpellCard(id)}
			},
			Label: "{T}, Remove a charge counter from this artifact: Add this artifact's last noted type and amount of mana. Spend this mana only to cast the last card exiled with this artifact.",
		}},
	})
}

// iceCauldronExile exiles the chosen card with the permission to cast it for
// as long as it stays exiled, and notes it as the last card exiled with the
// Cauldron. An on-disk key: never renamed, never reused.
var iceCauldronExile = effects.AskedPick("choose-cards/ice-cauldron-exile", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	id := r.Picked[0]
	if err := (effects.ExileWithPermission{Target: id, CastOnly: true, WhileExiled: true, Linked: true}).Apply(ctx); err != nil {
		return err
	}
	ctx.Game.SetPrefixedNoteForEffect(r.Source, "cauldron-card:", id.String())
	return nil
})
