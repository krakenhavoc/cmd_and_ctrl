package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Priest of Forgotten Gods — Creature — Human Cleric {1}{B}, 1/2
// (EDHREC rank 1101):
//
//	"{T}, Sacrifice two other creatures: Any number of target players
//	 each lose 2 life and sacrifice a creature of their choice. You add
//	 {B}{B} and draw a card."
//
// The aristocrats deck's edict engine. The cost is a tap plus a
// sacrifice clause with a count of two (#747, SacrificeN): the
// activator names exactly two creatures, and they leave as one
// simultaneous exit, so a Blood Artist paid in with a Goblin drains
// for both. The Priest taps, so CR 302.6 summoning sickness applies.
//
// "Any number of target players" is a zero-to-unbounded player clause
// (Stonespeaker Crystal's): nobody, everyone, or any mix, yourself
// included. On resolution each still-legal target loses 2 life and is
// asked to sacrifice a creature of their choice (a player with no
// creature is skipped, CR 701.21a); then the controller adds {B}{B}
// — it has targets, so it is not a mana ability (CR 605.1a) and the
// mana arrives on resolution — and draws. The target players'
// sacrifice prompts are queued in APNAP order (CR 101.4: the active
// player chooses first, then the others in turn order) and the mana
// and the draw are the run's continuation (#1019), so they land after
// the last opponent has answered, which is the printed order. They
// used to land first; nothing a player can see depended on it, but
// the sequence is observable and now it is right.
//
// Declared simplification, weaker than printed: "two OTHER creatures"
// is enforced by name (b03NotNamed, Warren Soultrader's posture),
// because a sacrifice clause's predicate never sees the source. A
// second Priest of Forgotten Gods, or a token copy of this one, cannot
// be fed to it.
func init() {
	Register(Spec{
		OracleID:     "2ad8ff62-d090-4835-9274-3b755ba0f8e6",
		Name:         "Priest of Forgotten Gods",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Another creature named Priest of Forgotten Gods can't be one of the two creatures sacrificed to its ability."},
		Activated: []ActivatedAbility{{
			Label: "{T}, Sacrifice two other creatures: Any number of target players each lose 2 life and sacrifice a creature. You add {B}{B} and draw a card.",
			Cost: Plus(TapCost(), SacrificeN(2, "two other creatures",
				Creature(), b03NotNamed("Priest of Forgotten Gods"))),
			Targets: TargetPlayer("any number of target players").WithCount(0, 0),
			Effect:  priestOfForgottenGodsEffect,
		}},
	})
}

// priestOfForgottenGodsEffect is the Priest's resolution, in printed
// order: each legal target player loses 2 life and sacrifices a
// creature of their choice, then the controller adds {B}{B} and draws.
func priestOfForgottenGodsEffect(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targeted := map[uuid.UUID]bool{}
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			targeted[t.ID] = true
		}
	}
	// APNAP order (CR 101.4), not the order the targets were named:
	// the sacrifice prompts are queued active player first.
	var players []uuid.UUID
	for _, p := range apnapPlayers(g) {
		if targeted[p] {
			players = append(players, p)
		}
	}
	for _, p := range players {
		if err := g.ChangePlayerLifeForEffect(item.SourceCardID, p, -2); err != nil {
			return err
		}
	}
	// One run over every target, so the controller's half waits for
	// the last of them. The answer is deliberately ignored: "you add
	// {B}{B} and draw a card" is not gated on anybody sacrificing
	// anything (ADR 0013 §5m item 5), and a table where nobody had a
	// creature still pays.
	return g.PlayersSacrificeThenForEffect(item.SourceCardID, players,
		sacrificeSpec("a creature", Creature()), "Sacrifice a creature",
		func(g *game.Game, _ game.PromptedSacrifices) error {
			if err := g.AddManaForEffect(item.Controller, item.SourceCardID, "{B}{B}"); err != nil {
				return err
			}
			return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
		})
}
