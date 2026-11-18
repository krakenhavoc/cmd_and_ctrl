package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jugan Defends the Temple // Remnant of the Rising Star — an Enchantment —
// Saga {2}{G} (Kamigawa: Neon Dynasty, S82):
//
//	I   — Create a 1/1 green Human Monk creature token with
//	      "{T}: Add {G}."
//	II  — Put a +1/+1 counter on each of up to two target creatures.
//	III — Exile this Saga, then return it to the battlefield
//	      transformed under your control.
//
// Back face (registered under "<oracle_id>#1"), Remnant of the Rising Star,
// an Enchantment Creature — Dragon Spirit 2/2:
//
//	Flying
//	Whenever another creature you control enters, you may pay {X}. When
//	you do, put X +1/+1 counters on that creature.
//	As long as you control five or more modified creatures, this creature
//	gets +5/+5 and has trample. (Equipment, Auras you control, and counters
//	are modifications.)
//
// The payment is Wildborn Preserver's MayPayX: X is chosen and paid on
// resolution, and the reflexive trigger (CR 603.12) carries X and the
// entering creature. The static is read live (CR 611.3a) off CR 700.9's
// "modified" (b07IsModified), counting the Remnant itself when it is one.
//
// No simplification.
func init() {
	const name = "Jugan Defends the Temple"
	Register(Spec{
		OracleID:     juganDefendsTheTempleOracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "create a 1/1 green Human Monk creature token with \"{T}: Add {G}.\""),
				Do(CreateToken{Template: neoB19HumanMonkToken(), N: 1})),
			ChapterTriggerTargeting(2, SagaChapterLabel(name, 2, "put a +1/+1 counter on each of up to two target creatures"),
				TargetCreature("up to two target creatures").WithCount(0, 2),
				putPlusOneCounterOnEachLegalTarget),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "exile this Saga, then return it transformed"),
				ChapterExileAndReturnTransformed),
		},
	})
	fiveModified := whileYouHave(neoB19FiveModifiedCreatures, selfCreature)
	grant := KeywordGrant(fiveModified, "trample")
	grant.Label = "As long as you control five or more modified creatures, this creature has trample."
	Register(Spec{
		OracleID:        juganDefendsTheTempleOracleID + "#1",
		Name:            "Remnant of the Rising Star",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:     "As long as you control five or more modified creatures, this creature gets +5/+5.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: fiveModified,
			Apply:     pumpBy(5, 5),
		}, grant},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b13AnotherCreatureYouControlEntered(ev, source, g)
			}, "Remnant of the Rising Star — you may pay {X}; when you do, put X +1/+1 counters on that creature",
				remnantOfTheRisingStarEnters),
		},
	})
}

// juganDefendsTheTempleOracleID is shared with the back face.
const juganDefendsTheTempleOracleID = "e70ac420-8540-4a69-820f-ae3e0274ca1a"

func remnantOfTheRisingStarEnters(g *game.Game, item *game.StackItem) error {
	entered := item.Trigger.Event.CardID
	return MayPayX{
		Cost:  "{X}",
		Label: "Remnant of the Rising Star",
		Buys:  "put X +1/+1 counters on that creature",
		Goal:  XAsHighAsYouCan,
		Then:  remnantOfTheRisingStarPaid,
		Carry: []uuid.UUID{entered},
	}.Apply(NewContext(g, item))
}

var remnantOfTheRisingStarPaid = PayUnlessBranches("pay-unless/remnant-of-the-rising-star-x-counters",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if r.Number <= 0 || len(r.Carry) == 0 {
			return nil
		}
		return ReflexiveTrigger{
			Label:  "Remnant of the Rising Star — put X +1/+1 counters on that creature",
			Body:   remnantOfTheRisingStarCountersBody,
			Cards:  []uuid.UUID{r.Carry[0]},
			Params: game.EffectParams{Amount: r.Number},
		}.Apply(ctx)
	}, nil)

// remnantOfTheRisingStarCountersBody is the reflexive trigger: X +1/+1
// counters on the creature that entered, when it is still on the battlefield.
var remnantOfTheRisingStarCountersBody = game.ReflexiveBody("remnant-of-the-rising-star/counters",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		ctx := NewContext(g, item)
		cards := ctx.PayloadCards()
		if p.Amount <= 0 || len(cards) == 0 || !onBattlefield(g, cards[0]) {
			return nil
		}
		return AddCounter{Target: cards[0], Kind: game.CounterPlusOne, N: p.Amount}.Apply(ctx)
	}, nil)
