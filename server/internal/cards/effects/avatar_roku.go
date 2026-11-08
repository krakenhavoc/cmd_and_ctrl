package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Avatar Roku — Legendary Creature — Avatar, 4/4 (colorless):
//
//	"Firebending 4 (Whenever this creature attacks, add {R}{R}{R}{R}.
//	 This mana lasts until end of combat.)
//	 {8}: Create a 4/4 red Dragon creature token with flying and
//	 firebending 4."
//
// The BACK face of The Legend of Roku ("<oracle_id>#1", ADR 0034),
// reached only through chapter III's exile-and-return
// (the_legend_of_roku.go).
//
// Firebending's mana lasts until end of combat (#3172): the four red
// mana carry the KeepManaUntilEndOfCombat mark, which the step-boundary
// sweep (game/mana_keep.go) honours through every step of the combat
// phase and drops at the first boundary outside it, so it can pay for
// instants and abilities in the declare blockers and damage steps and is
// gone by the postcombat main phase, as printed.
//
// The Dragon token DOES carry its own firebending 4 since ADR 0083
// (#1248): it is a catalog template under `token:dragon-firebending`,
// and its attack trigger is found through `game.CatalogKey`'s
// token-key fallback like a printed card's. Before that this card
// declared a second caveat for it — the "Triggered and static
// abilities on non-copy tokens" seam. Its mana is marked the
// same way.
func init() {
	Register(Spec{
		OracleID:     theLegendOfRokuOracleID + "#1",
		Name:         "Avatar Roku",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Avatar Roku — firebending 4", Do(firebending(4))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{8}: Create a 4/4 red Dragon creature token with flying and firebending 4.",
			Cost:    ManaCost("{8}"),
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Effect:  Do(CreateToken{Template: firebendingDragonToken(), N: 1}),
		}},
	})
}

// firebendingDragonToken is Avatar Roku's 4/4 red Dragon with flying
// and firebending 4.
func firebendingDragonToken() game.Card { return tokenFromCatalog(printedFirebendingDragonToken) }

// printedFirebendingDragonToken is that Dragon as PRINTED. Flying is
// a printed keyword and rides the Card as plain data; firebending is
// a triggered ability and rides the catalog slot, which is the whole
// distinction ADR 0083 draws.
func printedFirebendingDragonToken() tokenTemplate {
	return tokenTemplate{
		Slug: "dragon-firebending",
		Card: game.Card{
			Name:      "Dragon",
			TypeLine:  "Token Creature — Dragon",
			Power:     4,
			Toughness: 4,
			Colors:    []string{"R"},
			Keywords:  []string{"flying"},
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Dragon — firebending 4", Do(firebending(4))),
		},
		Text: "Flying\nFirebending 4 (Whenever this creature attacks, add {R}{R}{R}{R}. " +
			"This mana lasts until end of combat.)",
	}
}

// firebending is the whole effect of a "Firebending N" ability (CR
// 702.189a): add N red mana that lasts until end of combat, and then, as
// the ability finishes resolving, announce that its controller firebent
// (CR 702.189b, #5276; game.EventFirebend), which "whenever you firebend"
// (Avatar Aang) watches. N of zero or less adds nothing and still
// firebends: the ability resolved.
//
// Only a firebending ability's effect: red mana kept until end of combat
// that is not firebending (Avatar Roku, Firebender; Fire Lord Ozai) is
// redUntilEndOfCombat.
func firebending(n int) Firebending { return Firebending{N: n} }

// Firebending is firebending's Applier.
type Firebending struct{ N int }

// Apply adds the mana, then announces the firebend.
func (f Firebending) Apply(ctx *Context) error {
	if f.N > 0 {
		if err := redUntilEndOfCombat(f.N).Apply(ctx); err != nil {
			return err
		}
	}
	ctx.Game.NoteBendForEffect(game.EventFirebend, ctx.Controller(), ctx.Source())
	return nil
}

// redUntilEndOfCombat is "add N {R}. Until end of combat, you don't lose
// this mana as steps end", the mana half of firebending, for an effect
// that is not a firebending ability.
func redUntilEndOfCombat(n int) AddMana {
	return AddMana{
		Produced: game.RepeatMana("{R}", n),
		Riders:   []game.ManaSpendRider{KeepManaUntilEndOfCombat()},
	}
}
