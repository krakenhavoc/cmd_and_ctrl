package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Master, Formed Anew — Legendary Creature — Time Lord Rogue {U}{B}, 0/1
// (Doctor Who):
//
//	"Body Thief — When you cast this spell, you may exile a creature you
//	 control and put a takeover counter on it.
//	 You may have The Master enter as a copy of a creature card in exile with a
//	 takeover counter on it."
//
// The cast trigger resolves above the spell: its controller is asked which of
// their creatures to exile (none is fine), and a creature that really went
// gets a takeover counter in exile (CR 122.2: counters put on it there stay
// until it leaves). The entry copy is Body Double's, over the creature cards in
// exile that carry a takeover counter: any player's, so a second Master finds
// the first one's victims. The copy takes the card's printed values (CR 707.2);
// declining, or a pick gone by the time it is answered, leaves a 0/1 Master.
// Tokens cease to exist in exile, so only a card can ever be copied.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e7828284-543d-4f8b-9c5e-a8f082010398",
		Name:         "The Master, Formed Anew",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenYouCastThisSpell("Body Thief — you may exile a creature you control and put a takeover counter on it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ChoosePermanents{
						Question:   "The Master — exile a creature you control and put a takeover counter on it?",
						Candidates: effects.R34UpToOneCreatureYouControl,
						Keyed:      theMasterBodyThiefChosen,
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Replacements: []game.ReplacementEffect{
			effects.EntersAsCopyOf("The Master, Formed Anew", gapB85CreatureCardsInExileWithTakeoverCounter, nil),
		},
	})
}

// GapB85TakeoverCounter is the counter kind Body Thief puts on the exiled
// creature. A card-local kind: nothing else reads it.
const GapB85TakeoverCounter = "takeover"

// theMasterBodyThiefChosen exiles the creature picked and puts a takeover
// counter on it. A pick that has left the battlefield is not exiled.
var theMasterBodyThiefChosen effects.PermanentsThen

func init() {
	theMasterBodyThiefChosen = effects.RegisterPermanentsPick("prompt-run/the-master-formed-anew-body-thief", effects.PermanentsDef{
		Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
			for _, id := range picked.Cards() {
				if !effects.OnBattlefield(ctx.Game, id) {
					continue
				}
				if err := (effects.ExileTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
				if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
					continue
				}
				if err := (effects.AddCounter{Target: id, Kind: GapB85TakeoverCounter, N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

// gapB85CreatureCardsInExileWithTakeoverCounter is the copy's candidate list:
// every face-up creature card in exile with a takeover counter on it.
func gapB85CreatureCardsInExileWithTakeoverCounter(g *game.Game, _ uuid.UUID, _ uuid.UUID) []uuid.UUID {
	if g == nil || g.Exile == nil {
		return nil
	}
	var out []uuid.UUID
	for _, c := range g.Exile.Cards {
		if c.IsCreature() && !c.FaceDown && c.Counters[GapB85TakeoverCounter] > 0 {
			out = append(out, c.InstanceID)
		}
	}
	return out
}
