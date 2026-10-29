package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Seize the Spotlight — Sorcery {2}{R} (Edea steal-and-sac deck,
// #1565):
//
//	"Each opponent chooses fame or fortune. For each player who chose
//	 fame, gain control of a creature that player controls until end
//	 of turn. Untap those creatures and they gain haste until end of
//	 turn. For each player who chose fortune, you draw a card and
//	 create a Treasure token."
//
// Torment of Hailfire's per-opponent option pick (#568), asked one
// opponent at a time in APNAP order so each hears the choices before
// theirs (CR 101.4), then the two halves in printed order:
//
//   - **Fame.** For each player who chose fame, the CASTER chooses one
//     creature that player controls (a resolution-time pick over that
//     board, ChoosePermanents: it does not target, so hexproof does
//     not stop it). Each is then Act of Treason'd: gain control until
//     end of turn, untap, haste until end of turn. A player who chose
//     fame with no creature gives nothing.
//   - **Fortune.** One card and one Treasure per player who chose it.
//
// An opponent who leaves the game before answering is skipped, and the
// run goes on.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3d1b09a3-f151-4218-974b-02bb6159247a",
		Name:         "Seize the Spotlight",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return seizeTheSpotlightAsk(ctx, ctx.Opponents(), nil, 0)
		},
	})
	seizeTheSpotlightFortuneSteps = RegisterThenSteps("then-steps/seize-the-spotlight-fortune", seizeTheSpotlightFortuneStepList()...)
}

// seizeTheSpotlightAsk asks the next opponent, accumulating the fame
// seats and the fortune count, and runs the effect once everyone has
// answered.
func seizeTheSpotlightAsk(ctx *Context, remaining, fame []uuid.UUID, fortune int) error {
	for len(remaining) > 0 {
		opp := remaining[0]
		rest := remaining[1:]
		if p := ctx.Game.PlayerByIDForEffect(opp); p == nil || p.Eliminated {
			remaining = rest
			continue
		}
		return PickOption{
			Player:   opp,
			Question: "Seize the Spotlight — choose fame or fortune",
			Options: []game.ChoiceOption{
				{Label: "Fame — its caster gains control of a creature you control until end of turn"},
				{Label: "Fortune — its caster draws a card and creates a Treasure token"},
			},
			ThenKey:  seizeTheSpotlightAnswered,
			WithItem: true,
			// The seats still to ask, behind the fame seats so far; the
			// fortunes so far, and how many of the carry are fame seats.
			Carry: append(append([]uuid.UUID(nil), fame...), rest...),
			Nums:  []int{fortune, len(fame)},
		}.Apply(ctx)
	}
	return seizeTheSpotlightResolve(ctx, fame, fortune)
}

// seizeTheSpotlightAnswered is one opponent's answer, keyed (#4253): the seat
// that answered is the chooser, and the fame seats so far, the seats still to
// ask, the fortunes so far and how many of the carry are fame seats are its
// plain data. An on-disk key: never renamed, never reused. Registered in
// init because the walk asks again from inside it.
var seizeTheSpotlightAnswered game.OptionPickThen

func init() {
	seizeTheSpotlightAnswered = OptionPickThen("option-pick/seize-the-spotlight-answer", func(ctx *Context, r game.OptionPicked) error {
		if len(r.Nums) < 2 || r.Nums[1] < 0 || r.Nums[1] > len(r.Carry) {
			return nil
		}
		fortune := r.Nums[0]
		fame, rest := r.Carry[:r.Nums[1]], r.Carry[r.Nums[1]:]
		// One question asked per answer: the arms are exclusive.
		nextFame, nextFortune := fame, fortune
		switch r.Index {
		case 0:
			nextFame = append(append([]uuid.UUID(nil), fame...), r.Chooser)
		case 1:
			nextFortune++
		}
		return seizeTheSpotlightAsk(ctx, rest, nextFame, nextFortune)
	})
}

// seizeTheSpotlightFortune is the fortune half: a card and a Treasure,
// fortune times, as continuation steps (#4162, CR 608.2c). The Treasure
// waits for a prompt the draw asks (a dredge offer), and the next card
// for one the Treasure asks.
func seizeTheSpotlightFortune(ctx *Context, fortune int) error {
	if fortune <= 0 {
		return nil
	}
	ctx.RecordStep(seizeTheSpotlightRecord, nil, fortune)
	return seizeTheSpotlightFortuneSteps.Run(ctx)
}

// seizeTheSpotlightRecord is how many fortunes are still owed.
const seizeTheSpotlightRecord = "seize-the-spotlight/fortune"

// seizeTheSpotlightFortuneSteps is registered in init, because its last
// step runs it again for the next fortune.
var seizeTheSpotlightFortuneSteps ThenSteps

func seizeTheSpotlightFortuneStepList() []Applier {
	return []Applier{
		DrawCards{N: 1},
		Instruction(func(ctx *Context) error {
			return CreateToken{Template: TreasureToken(), N: 1}.Apply(ctx)
		}),
		Instruction(func(ctx *Context) error {
			return seizeTheSpotlightFortune(ctx, ctx.StepRecord(seizeTheSpotlightRecord).N()-1)
		}),
	}
}

// seizeTheSpotlightResolve runs the fame half, then the fortune half.
func seizeTheSpotlightResolve(ctx *Context, fame []uuid.UUID, fortune int) error {
	if len(fame) == 0 {
		return seizeTheSpotlightFortune(ctx, fortune)
	}
	return ChoosePermanents{
		Question: "Seize the Spotlight — gain control of a creature that player controls",
		Of:       fame,
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == of && c.IsCreature() {
					out = append(out, c.InstanceID)
				}
			}
			if len(out) == 0 {
				return nil, 0, 0
			}
			return out, 1, 1
		},
		Keyed: seizeTheSpotlightChosen,
		Nums:  []int{fortune},
	}.Apply(ctx)
}

// seizeTheSpotlightChosen is gaining control of each creature chosen, then
// the fortune half, keyed (#4290): how many fortunes are owed is the run's
// number.
var seizeTheSpotlightChosen = RegisterPermanentsPick("prompt-run/seize-the-spotlight-gain-control", PermanentsDef{
	ThenRun: func(ctx *Context, picked game.PromptedPicks, r game.RunDone) error {
		for _, id := range picked.Cards() {
			if err := (GainControl{
				Target:   id,
				Duration: DurationUntilEndOfTurn(ctx),
				Label:    "Seize the Spotlight — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (UntapTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			if err := (GrantKeywordUntilEOT{
				Target:   id,
				Keywords: []string{"haste"},
				Label:    "Seize the Spotlight — haste until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
		}
		fortune := 0
		if len(r.Nums) > 0 {
			fortune = r.Nums[0]
		}
		return seizeTheSpotlightFortune(ctx, fortune)
	},
})
