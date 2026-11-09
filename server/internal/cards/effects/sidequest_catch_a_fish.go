package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sidequest: Catch a Fish // Cooking Campsite (Final Fantasy, #4462).
// Front, Enchantment {2}{W}:
//
//	"At the beginning of your upkeep, look at the top card of your
//	 library. If it's an artifact or creature card, you may reveal it and
//	 put it into your hand. If you put a card into your hand this way,
//	 create a Food token and transform this enchantment."
//
// Back, Land:
//
//	"{T}: Add {W}.
//	 {3}, {T}, Sacrifice an artifact: Put a +1/+1 counter on each
//	 creature you control. Activate only as a sorcery."
//
// The look is Bison Whistle's (LookAtTopOfLibraryForEffect), the offer a
// MayChoice that is asked only for an artifact or creature card (the
// ruling: you needn't reveal it; it stays on top). Revealing it and
// moving it to the hand is Bison Whistle's take; the Food and the
// transform follow only if the card really reached the hand. The
// transform is the in-place verb on this enchantment (a land on the back
// face), so a Sidequest that left the battlefield first is not turned
// over.
//
// Cooking Campsite is Abandoned Air Temple's mana ability and counter
// spread behind a sacrifice cost, at sorcery speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     sidequestCatchAFishOracleID,
		Name:         "Sidequest: Catch a Fish",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Sidequest: Catch a Fish — look at the top card of your library; if it's an artifact or creature card, you may reveal it and put it into your hand, then create a Food token and transform this enchantment",
				Do(Instruction(sidequestCatchAFishLook))),
		},
	})
	Register(Spec{
		OracleID:     sidequestCatchAFishOracleID + "#1",
		Name:         "Cooking Campsite",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "{T}: Add {W}.",
		}},
		Activated: []ActivatedAbility{{
			Label:        "{3}, {T}, Sacrifice an artifact: Put a +1/+1 counter on each creature you control. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{3}"), TapCost(), SacrificeN(1, "an artifact", Artifact())),
			SorcerySpeed: true,
			Effect:       b11PutCounterOnEachCreatureYouControl,
		}},
	})
}

// sidequestCatchAFishOracleID is shared by both faces.
const sidequestCatchAFishOracleID = "bd7c328e-0380-46f8-bb85-7bf4e201b7ac"

// sidequestCatchAFishLook is the upkeep ability.
func sidequestCatchAFishLook(ctx *Context) error {
	looked := ctx.Game.LookAtTopOfLibraryForEffect(ctx.Controller(), 1)
	if len(looked) == 0 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(looked[0])
	if !ok || !(c.IsCreature() || c.IsArtifact()) {
		return nil
	}
	return MayChoice{
		Player:   ctx.Controller(),
		Question: "Sidequest: Catch a Fish — reveal " + c.Name + " and put it into your hand?",
		YesLabel: "Put it into your hand",
		NoLabel:  "Leave it on top",
		Then:     sidequestCatchAFishTake,
		Carry:    []uuid.UUID{looked[0]},
	}.Apply(ctx)
}

// sidequestCatchAFishTake reveals the card, puts it into the hand, and
// if it got there makes the Food and transforms the enchantment.
var sidequestCatchAFishTake = MayYes("fin-b1/sidequest-catch-a-fish-take", func(ctx *Context, r game.ConfirmAnswer) error {
	id := carryAt(r, 0)
	if !bisonWhistleStillInLibrary(ctx.Game, id) {
		return nil
	}
	ctx.Game.RevealTopOfLibraryForEffect(r.Chooser, ctx.Source(), 1, "Sidequest: Catch a Fish — reveal the top card")
	if err := ctx.Game.BounceToHandForEffect(id); err != nil {
		return err
	}
	if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneHand {
		return nil
	}
	if err := (CreateToken{Template: FoodToken(), N: 1}).Apply(ctx); err != nil {
		return err
	}
	return Transform{Target: ctx.Source()}.Apply(ctx)
})
