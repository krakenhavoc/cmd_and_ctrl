package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Prowl, Stoic Strategist // Prowl, Pursuit Vehicle (#2651, #3195, ADR
// 0152) — a More Than Meets the Eye card.
//
// Front face, Legendary Artifact Creature — Robot {3}{W}, 3/3:
//
//	"More Than Meets the Eye {2}{W}
//	 Whenever Prowl attacks, exile up to one other target tapped
//	 creature or Vehicle. For as long as that card remains exiled, its
//	 owner may play it.
//	 Whenever a player plays a card exiled with Prowl, you draw a card
//	 and convert Prowl."
//
// Back face, Legendary Artifact — Vehicle, 2/3:
//
//	"Living metal
//	 Whenever another creature or Vehicle you control enters, put a
//	 +1/+1 counter on Prowl. If this is the second time this ability
//	 has resolved this turn, convert Prowl."
//
// The attack trigger's exile is LINKED (ExileWithPermission.Linked): the
// card lands with Card.ExiledWith naming this Prowl object and
// Card.ExiledWithAbility naming the attack trigger's catalog row, and
// its owner gets a play permission that lasts while it stays in exile
// (a land too: "play", not "cast"). The second trigger is the other half
// of the CR 607.2a pair. The play moves the card out of exile, which
// clears its link (CR 400.7), so the play's own event carries the link
// it had (Event.ExiledWith, #2651) and the trigger reads that: only a
// card this Prowl's attack trigger exiled counts, not one another
// ability or another Prowl exiled, and not one an earlier Prowl object
// exiled before it left and came back.
//
// The back face's count is Elrond, Lord of Rivendell's: this object's
// resolutions of the ability this turn, the resolving one included
// (Game.ResolvedThisTurn), read after the counter is placed.
//
// No simplification.
const prowlOracleID = "8dfd9baf-c3cc-4d2f-a0b0-b323b6480562"

const (
	prowlExileLabel = "Prowl, Stoic Strategist — exile up to one other target tapped creature or Vehicle; its owner may play it while it remains exiled"
	prowlPlayLabel  = "Prowl, Stoic Strategist — a card exiled with Prowl was played: draw a card and convert Prowl"
	prowlBackLabel  = "Prowl, Pursuit Vehicle — put a +1/+1 counter on Prowl; the second time this turn, convert Prowl"
)

func init() {
	Register(Spec{
		OracleID:         prowlOracleID,
		Name:             "Prowl, Stoic Strategist",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{MoreThanMeetsTheEye("{2}{W}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisAttacks(prowlExileLabel, prowlExile),
				Another(TargetPermanent("up to one other target tapped creature or Vehicle",
					Or(Creature(), HasSubtype("Vehicle")), b751Tapped())).WithCount(0, 1)),
			OnAny([]game.EventKind{game.EventCast, game.EventZoneMove}, prowlPlayedACardExiledWithIt,
				prowlPlayLabel, prowlDrawAndConvert),
		},
	})
	Register(Spec{
		OracleID:     prowlOracleID + "#1",
		Name:         "Prowl, Pursuit Vehicle",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{LivingMetal()},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, anotherCreatureOrVehicleYouControlEntered, prowlBackLabel, prowlBackCounter),
		},
	})
}

// prowlExile exiles the chosen target, if it is still legal (CR 608.2b),
// linked to this Prowl and its attack trigger, with a play permission
// for its owner while it stays exiled.
func prowlExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return ExileWithPermission{Target: t.ID, WhileExiled: true, Linked: true}.Apply(ctx)
	}
	return nil
}

// prowlPlayedACardExiledWithIt is "whenever a player plays a card exiled
// with Prowl": a spell cast or a land played from exile whose link names
// this Prowl object and its attack trigger.
func prowlPlayedACardExiledWithIt(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return game.PlayedCardExiledWithRow(ev,
		game.PermissionCardRef{ID: source.InstanceID, Epoch: source.ObjectEpoch},
		game.AbilitySlotTriggered, prowlExileLabel)
}

// prowlDrawAndConvert is "you draw a card and convert Prowl".
func prowlDrawAndConvert(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
		return err
	}
	return ConvertThis{}.Apply(ctx)
}

// anotherCreatureOrVehicleYouControlEntered is "whenever another
// creature or Vehicle you control enters".
func anotherCreatureOrVehicleYouControlEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.CardID == source.InstanceID {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Controller == source.Controller && (c.IsCreature() || c.HasSubtype("Vehicle"))
}

// prowlBackCounter puts a +1/+1 counter on Prowl and, the second time
// this object's ability resolves this turn, converts it.
func prowlBackCounter(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	if g.ResolvedThisTurn(item.SourceCardID, prowlBackLabel) != 2 {
		return nil
	}
	return ConvertThis{}.Apply(ctx)
}
