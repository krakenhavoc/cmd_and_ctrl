package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elrond of the White Council — Legendary Creature — Elf Noble
// {3}{G}{U}, 3/3:
//
//	"Secret council — When Elrond enters, each player secretly votes
//	 for fellowship or aid, then those votes are revealed. For each
//	 fellowship vote, the voter chooses a creature they control. You
//	 gain control of each creature chosen this way, and they gain 'This
//	 creature can't attack its owner.' Then for each aid vote, put a
//	 +1/+1 counter on each creature you control."
//
// A secret vote (ADR 0180). Once the votes are revealed, each player
// who voted fellowship chooses one creature they control per fellowship
// vote, in the order they voted (a voter with fewer creatures chooses
// all they have). No creature moves while players choose; then Elrond's
// controller gains control of every chosen creature still on the
// battlefield, for good (CR 611.2a: no duration is printed), and each
// gains "This creature can't attack its owner" (ADR 0180 decision 6,
// the owner read live as attacks are declared). Then each creature
// Elrond's controller controls, the new ones included, gets a +1/+1
// counter per aid vote.
//
// A fellowship vote by Elrond's controller makes them choose one of
// their own creatures, which changes nothing they control, so their bot
// votes aid; an opponent's bot votes fellowship, giving up one creature
// of its choosing rather than growing Elrond's whole board.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5ae4a172-b846-4ec2-8178-a8b762134bfe",
		Name:         "Elrond of the White Council",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Elrond of the White Council — each player secretly votes for fellowship or aid",
				func(g *game.Game, item *game.StackItem) error {
					return Vote{
						Question:      "Elrond of the White Council — secretly vote for fellowship or aid",
						Words:         []string{"fellowship", "aid"},
						Secret:        true,
						ForController: []int{0, 2},
						ForOpponents:  []int{1, 0},
						Then:          elrondVoted,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

var elrondVoted = VoteResultThen("vote/elrond-of-the-white-council", func(ctx *Context, r game.VoteResult) error {
	owed := map[uuid.UUID]int{}
	var voters []uuid.UUID
	for _, v := range r.VotersFor(0) {
		if owed[v] == 0 {
			voters = append(voters, v)
		}
		owed[v]++
	}
	return elrondAskFellowship(ctx, voters, owed, nil, r.Votes(1))
})

// elrondAskFellowship asks the first fellowship voter for their
// creatures, then the next, and once every voter has chosen takes
// them all and hands out the aid counters.
func elrondAskFellowship(ctx *Context, voters []uuid.UUID, owed map[uuid.UUID]int, chosen []uuid.UUID, aid int) error {
	for i, voter := range voters {
		if !voterStillIn(ctx.Game, voter) || len(creaturesControlledByPlayer(ctx.Game, voter)) == 0 {
			continue
		}
		rest := append([]uuid.UUID(nil), voters[i+1:]...)
		so := append([]uuid.UUID(nil), chosen...)
		n := owed[voter]
		return ChoosePermanents{
			Player:   voter,
			Of:       []uuid.UUID{voter},
			Question: "Elrond of the White Council — choose a creature you control for each of your fellowship votes; Elrond's controller gains control of it",
			Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
				ids := creaturesControlledByPlayer(g, of)
				k := min(n, len(ids))
				return ids, k, k
			},
			Keyed: elrondChosen,
			Carry: packIDLists(rest, so),
			Nums:  elrondNums(aid, rest, owed),
		}.Apply(ctx)
	}
	return elrondTakeAndAid(ctx, chosen, aid)
}

// elrondNums is the run's numbers: the aid counters, then how many each
// voter still to ask owes, in their order.
func elrondNums(aid int, rest []uuid.UUID, owed map[uuid.UUID]int) []int {
	out := []int{aid}
	for _, v := range rest {
		out = append(out, owed[v])
	}
	return out
}

// elrondChosen is one fellowship voter's pick, then the next voter's,
// keyed (#4290). The carry is the voters still to ask, then the creatures
// chosen so far; the numbers are elrondNums. Registered in init because the
// answer asks the next voter, which names the key again.
var elrondChosen PermanentsThen

func init() {
	elrondChosen = RegisterPermanentsPick("prompt-run/elrond-of-the-white-council-fellowship", PermanentsDef{
		ThenRun: func(ctx *Context, picked game.PromptedPicks, r game.RunDone) error {
			rest, so := unpackIDPair(r.Carry)
			aid := 0
			owed := map[uuid.UUID]int{}
			if len(r.Nums) > 0 {
				aid = r.Nums[0]
				for i, v := range rest {
					if i+1 < len(r.Nums) {
						owed[v] = r.Nums[i+1]
					}
				}
			}
			return elrondAskFellowship(ctx, rest, owed, append(so, picked.Cards()...), aid)
		},
	})
}

// elrondTakeAndAid gains control of each chosen creature still on the
// battlefield, gives it "can't attack its owner", then puts `aid` +1/+1
// counters on each creature the controller controls.
func elrondTakeAndAid(ctx *Context, chosen []uuid.UUID, aid int) error {
	for _, id := range chosen {
		if !onBattlefield(ctx.Game, id) {
			continue
		}
		if err := (GainControl{Target: id, Duration: game.IndefiniteDuration(), Label: "Elrond of the White Council — gain control"}).Apply(ctx); err != nil {
			return err
		}
		if err := (ScopedEffectFor{
			Target:   id,
			Mods:     []game.Mod{game.CantAttackOwnerMod()},
			Duration: game.IndefiniteDuration(),
			Label:    "Elrond of the White Council — this creature can't attack its owner",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	if aid <= 0 {
		return nil
	}
	// The creatures just taken are "creatures you control" now: settle
	// the control change before reading who controls what.
	ctx.Game.RecomputeLayersIfStaleLocked()
	for _, id := range creaturesControlledByPlayer(ctx.Game, ctx.Controller()) {
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: aid}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
