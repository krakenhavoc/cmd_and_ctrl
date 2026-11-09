package effects

import (
	"errors"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scavenger's Talent — Enchantment — Class {B} (Bloomburrow, #4467):
//
//	"(Gain the next level as a sorcery to add its ability.)
//	 Whenever one or more creatures you control die, create a Food token.
//	 This ability triggers only once each turn.
//	 {1}{B}: Level 2
//	 Whenever you sacrifice a permanent, target player mills two cards.
//	 {2}{B}: Level 3
//	 At the beginning of your end step, you may sacrifice three other
//	 nonland permanents. If you do, return a creature card from your
//	 graveyard to the battlefield with a finality counter on it."
//
// ADR 0071 Class machinery (Bandit's Talent): LevelUp for each level, the
// level-2 and level-3 triggers gated with AtLevel (CR 716.2a).
//
//   - Level 1 is Homicide Investigator's shape over any creature (a token
//     included): one batch trigger per death event (CR 603.2c), a
//     leaves-the-battlefield ability that looks back, and
//     TriggersOnlyOnceEachTurn.
//   - Level 2 watches EventSacrifice, which fires per permanent while it is
//     still on the battlefield, so it triggers once for each permanent
//     sacrificed, the Talent itself included (the rulings). Any cost or
//     effect that sacrifices counts, and its trigger resolves before the
//     spell or ability that was paid for. The mill's target is a player.
//   - Level 3 is a may-pay whose payment is "sacrifice three other nonland
//     permanents" (game.PayAction: Count 3, Nonland, Except the Talent), so
//     an ineligible board is never asked and "if you do" is false. The
//     paid branch then asks for a creature card in your graveyard, the
//     sacrificed creatures included (the rulings), and returns it with a
//     finality counter. The sacrifices announce EventSacrifice like any
//     other, so level 2 sees them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aefea339-8a0d-4531-8a62-afaecc88d078",
		Name:         "Scavenger's Talent",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggersOnlyOnceEachTurn(OncePerBatch(On(game.EventLTB, scavengersTalentCreatureYouControlDied,
				"Scavenger's Talent — create a Food token",
				Do(CreateToken{Template: FoodToken(), N: 1})))),
			AtLevel(2, Targeting(On(game.EventSacrifice, scavengersTalentYouSacrificed,
				"Scavenger's Talent — target player mills two cards",
				targetPlayerMillsTwo), TargetPlayer("target player"))),
			AtLevel(3, AtYourEndStep("Scavenger's Talent — you may sacrifice three other nonland permanents; if you do, return a creature card from your graveyard with a finality counter",
				func(g *game.Game, item *game.StackItem) error {
					pay := SacrificePayment(3, "other nonland permanent", "other nonland permanents", game.PermanentQuery{})
					pay.Action.Nonland = true
					except := item.SourceCardID
					pay.Action.Except = &except
					pay.Words = "Sacrifice three other nonland permanents"
					return MayPay{
						Chooser:  item.Controller,
						Pay:      pay,
						Question: "Scavenger's Talent — sacrifice three other nonland permanents to return a creature card from your graveyard with a finality counter?",
						Then:     scavengersTalentPaid,
					}.Apply(NewContext(g, item))
				})),
		},
		Activated: []ActivatedAbility{
			LevelUp(2, ManaCost("{1}{B}")),
			LevelUp(3, ManaCost("{2}{B}")),
		},
	})
}

// scavengersTalentCreatureYouControlDied is one death in the batch: a
// creature (token or not) the Talent's controller controlled as it died.
func scavengersTalentCreatureYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	return ok && leftUnderControlOf(ev, dead) == source.Controller
}

// scavengersTalentYouSacrificed is "whenever you sacrifice a permanent".
func scavengersTalentYouSacrificed(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventSacrifice && ev.Actor == source.Controller && ev.CardID != uuid.Nil
}

// scavengersTalentPaid is "if you do": the pick of the creature card.
// An on-disk key.
var scavengersTalentPaid = PayUnlessBranches("pay-unless/scavengers-talent-return",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		g, you := ctx.Game, ctx.Controller()
		var candidates []uuid.UUID
		if p := g.PlayerByIDForEffect(you); p != nil && p.Graveyard != nil {
			for _, c := range p.Graveyard.Cards {
				if c.IsCreature() {
					candidates = append(candidates, c.InstanceID)
				}
			}
		}
		if len(candidates) == 0 {
			return nil
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    you,
			FromPlayer: you,
			Source:     ctx.Source(),
			Question:   "Scavenger's Talent — return a creature card from your graveyard to the battlefield with a finality counter on it",
			Cards:      candidates,
			Min:        1,
			Max:        1,
			LiveZone:   game.ZoneGraveyard,
			ThenKey:    scavengersTalentReturned,
			Asked:      game.NewAskedItem(ctx.Item),
		})
		return nil
	}, nil)

// scavengersTalentReturned puts the chosen creature card onto the
// battlefield with a finality counter. An on-disk key.
var scavengersTalentReturned = askedPick("choose-cards/scavengers-talent-return", func(ctx *Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		_, err := ctx.Game.ReturnFromGraveyardWithCountersForEffect(id, uuid.Nil, false,
			map[string]int{game.CounterFinality: 1})
		if err != nil && !errors.Is(err, game.ErrCardNotFound) {
			return err
		}
	}
	return nil
})
