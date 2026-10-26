package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zenos yae Galvus // Shinryu, Transcendent Rival — Legendary Creature —
// Human Noble Warrior {3}{B}{B}, 4/4, transforming into a Legendary
// Creature — Dragon, 8/8 (the Squall deck request, #3918):
//
//	"My First Friend — When Zenos yae Galvus enters, choose a creature
//	 an opponent controls. Until end of turn, creatures other than Zenos
//	 yae Galvus and the chosen creature get -2/-2.
//	 When the chosen creature leaves the battlefield, transform Zenos
//	 yae Galvus."
//	// "Flying
//	 As this creature transforms into Shinryu, choose an opponent.
//	 Burning Chains — When the chosen player loses the game, you win the
//	 game."
//
// MY FIRST FRIEND chooses as the trigger resolves; it does not target,
// so a hexproof creature can be chosen. The choice is stored on Zenos as
// an OBJECT (Card.ChosenPermanent, with its CR 400.7 epoch), and the
// -2/-2 lands on the creatures on the battlefield then (CR 611.2c),
// every one but Zenos and the chosen creature. With no creature to
// choose, every creature but Zenos gets -2/-2 and nothing is chosen.
//
// THE TRANSFORM TRIGGER watches the chosen object leave, matched by its
// last-known epoch, so the same card returning and leaving again is a
// new object nobody chose. Zenos leaving clears its choice (the chosen
// permanent's lifecycle), so a Zenos that left and came back waits on
// its own new choice. TransformThis honours CR 701.27f.
//
// SHINRYU chooses its opponent as it transforms (AsTransformsInto, the
// Sephiroth shape), stored as Card.ChosenPlayer. A Shinryu that enters
// on its back face never transformed and chooses nobody. Burning Chains
// watches the chosen player leaving the game for any reason a player
// loses (CR 104.3, conceding included) and, on resolution, its
// controller wins (CR 104.2b), unless something says they can't.
//
// No simplification.
const zenosYaeGalvusOracleID = "8f52fc65-a914-421a-9426-21eb520d55d5"

func init() {
	Register(Spec{
		OracleID:     zenosYaeGalvusOracleID,
		Name:         "Zenos yae Galvus",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Zenos yae Galvus — My First Friend: choose a creature an opponent controls; the others get -2/-2",
				zenosMyFirstFriend),
			On(game.EventLTB, zenosChosenCreatureLeft, "Zenos yae Galvus — transform it", Do(TransformThis{})),
		},
	})

	Register(Spec{
		OracleID:        zenosYaeGalvusOracleID + "#1",
		Name:            "Shinryu, Transcendent Rival",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		AsTransformsInto: func(card *game.Card, ctx *Context) error {
			ctx.Game.QueueChoosePlayerAsEntersForEffect(card.Controller, card.InstanceID,
				"Shinryu, Transcendent Rival — choose an opponent",
				Opponents.seats(ctx, card.Controller))
			return nil
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventPlayerEliminated, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.ChosenPlayer != uuid.Nil && ev.Actor == source.ChosenPlayer
			}, "Shinryu — Burning Chains: you win the game", Do(WinTheGame{})),
		},
	})
}

// zenosMyFirstFriend asks for the creature, then shrinks the rest.
func zenosMyFirstFriend(g *game.Game, item *game.StackItem) error {
	zenos := item.SourceCardID
	var candidates []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.Controller != item.Controller {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return zenosShrinkOthers(NewContext(g, item), zenos, uuid.Nil)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   zenos,
		Question: "Zenos yae Galvus — My First Friend: choose a creature an opponent controls",
		Cards:    candidates,
		Min:      1,
		Max:      1,
		ThenKey:  zenosChose,
		LiveZone: game.ZoneBattlefield,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// zenosChose is the choice of the creature to spare, keyed (#4253): Zenos
// is the asking card and the item rides the prompt as data.
var zenosChose = askedPick("choose-cards/zenos-yae-galvus-first-friend", func(ctx *Context, r game.CardsChosen) error {
	g, zenos := ctx.Game, r.Source
	chosen := uuid.Nil
	if len(r.Picked) == 1 {
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.InstanceID == r.Picked[0] {
				chosen = c.InstanceID
				g.SetChosenPermanentForEffect(zenos, game.ObjectRef{ID: chosen, Epoch: c.ObjectEpoch})
				break
			}
		}
	}
	return zenosShrinkOthers(ctx, zenos, chosen)
})

// zenosShrinkOthers is "until end of turn, creatures other than Zenos
// and the chosen creature get -2/-2", the set fixed now.
func zenosShrinkOthers(ctx *Context, zenos, chosen uuid.UUID) error {
	return BoostUntilEOT{
		Match: And(Creature(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
			return c.InstanceID != zenos && c.InstanceID != chosen
		}),
		Power:     -2,
		Toughness: -2,
		Label:     "Zenos yae Galvus — My First Friend: -2/-2",
	}.Apply(ctx)
}

// zenosChosenCreatureLeft is "when the chosen creature leaves the
// battlefield": the object stored on this Zenos, by its epoch.
func zenosChosenCreatureLeft(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	ref := source.ChosenPermanent
	if ev.Kind != game.EventLTB || ref.ID == uuid.Nil || ev.CardID != ref.ID {
		return false
	}
	info, ok := g.LastKnownPermanentForEffect(ev.CardID)
	return !ok || info.Epoch == ref.Epoch
}
