package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// venture.go is the card vocabulary of "venture into the dungeon" (CR
// 701.49), ADR 0203's amendment of 2026-10-19 (#4100). The engine runs
// the keyword action (game.VentureForEffect): entering a dungeon,
// advancing along an arrow, completing one from its last room. A plain
// venture chooses among every dungeon not marked VentureIntoOnly — Lost
// Mine of Phandelver, Dungeon of the Mad Mage and Tomb of Annihilation
// (dungeons_forgotten_realms.go) — so a card only has to say it:
//
//	VentureIntoTheDungeon{}           "venture into the dungeon"
//	Venture                           the same, as a whole Effect
//	Venturing(rest...)                "venture into the dungeon and <rest>"
//	WhenThisEntersVenture(name)       "When this <permanent> enters, venture into the dungeon."
//	PumpWhileCompletedADungeon,
//	KeywordWhileCompletedADungeon     "as long as you've completed a dungeon, …"
//	IfVentureCond(cond, trigger)      an intervening "if" these cards print (CR 603.4)

// VentureIntoTheDungeon is "<player> ventures into the dungeon" (CR
// 701.49a). Player defaults to the effect's controller. A player who has
// left the game ventures nowhere (CR 800.4a; the engine checks).
type VentureIntoTheDungeon struct {
	Player uuid.UUID
}

func (v VentureIntoTheDungeon) Apply(ctx *Context) error {
	player := v.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	return ctx.Game.VentureForEffect(player, ctx.Source(), "")
}

// Venture is "venture into the dungeon" as a whole Effect, for a
// trigger or an activated ability whose text is only that.
var Venture = Do(VentureIntoTheDungeon{})

// ventureLabel is the stack label of a card's venture ability.
func ventureLabel(name string) string { return name + " — venture into the dungeon" }

// WhenThisEntersVenture is "When this <permanent> enters, venture into
// the dungeon."
func WhenThisEntersVenture(name string) game.TriggeredAbility {
	return WhenThisEnters(ventureLabel(name), Venture)
}

// ventureCond is a CR 603.4 intervening "if" these cards print, as a
// function of the game and the ability's controller. It is both halves
// of the condition: read as the ability triggers (IfVentureCond's
// AppliesTo) and as it resolves (Holds). It declares no footprint.
type ventureCond func(g *game.Game, you uuid.UUID) bool

func (c ventureCond) Holds(ctx *Context) bool { return c(ctx.Game, ctx.Controller()) }

func (ventureCond) condition() (game.FootprintStep, bool) { return game.FootprintStep{}, false }

var (
	// CondYouCompletedADungeon is "if you've completed a dungeon".
	CondYouCompletedADungeon = ventureCond(YouCompletedADungeon)
	// condYouHaventCompletedTomb is "if you haven't completed Tomb of
	// Annihilation" (Acererak the Archlich).
	condYouHaventCompletedTomb = ventureCond(func(g *game.Game, you uuid.UUID) bool {
		return !game.CompletedDungeon(g, you, TombOfAnnihilationKey)
	})
	// condACreatureDiedThisTurn is "if a creature died this turn": any
	// creature, any controller (Zombie Ogre).
	condACreatureDiedThisTurn = ventureCond(func(g *game.Game, _ uuid.UUID) bool {
		return g.TurnTally.CreaturesDied > 0
	})
	// condYouHadALandEnter is "if you had a land enter the battlefield
	// under your control this turn" (Wandering Troubadour).
	condYouHadALandEnter = ventureCond(func(g *game.Game, you uuid.UUID) bool {
		return g.TurnTallyFor(you).LandsEntered > 0
	})
)

// IfVentureCond narrows a trigger to the intervening "if" `cond` (CR
// 603.4): it triggers only while the condition holds for its
// controller, and does nothing if it no longer holds as it resolves.
func IfVentureCond(cond ventureCond, t game.TriggeredAbility) game.TriggeredAbility {
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return source != nil && cond(g, source.Controller) && when(ev, source, lki, g)
	}
	return InterveningIf(cond, t)
}

// completedADungeonSource reports whether the controller of `source`
// has completed a dungeon: the condition a "as long as you've completed
// a dungeon" static reads. The layer pass bumps on EventDungeonCompleted
// (ADR 0203 §3), so a static gated on it is re-read the moment a
// dungeon is completed.
func completedADungeonSource(g *game.Game, source *game.Card) bool {
	return source != nil && YouCompletedADungeon(g, source.Controller)
}

// whileCompletedADungeon narrows a static's AppliesTo to "as long as
// you've completed a dungeon".
func whileCompletedADungeon(applies func(target *game.Card, g *game.Game, source *game.Card) bool) func(target *game.Card, g *game.Game, source *game.Card) bool {
	return func(target *game.Card, g *game.Game, source *game.Card) bool {
		return completedADungeonSource(g, source) && applies(target, g, source)
	}
}

// PumpWhileCompletedADungeon is "<these permanents> get +p/+t as long
// as you've completed a dungeon": a layer 7c modification (Cloister
// Gargoyle's +3/+0, Nadaar's anthem).
func PumpWhileCompletedADungeon(applies func(target *game.Card, g *game.Game, source *game.Card) bool, power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Label:     "Gives " + signedPT(power, toughness) + " as long as you've completed a dungeon.",
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: whileCompletedADungeon(applies),
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power += power
			c.Toughness += toughness
		},
	}
}

// KeywordWhileCompletedADungeon is "<these permanents> have <keyword>
// as long as you've completed a dungeon": a layer 6 grant (Gloom
// Stalker's double strike, Cloister Gargoyle's flying).
func KeywordWhileCompletedADungeon(applies func(target *game.Card, g *game.Game, source *game.Card) bool, keyword string) game.StaticAbility {
	ab := KeywordGrant(whileCompletedADungeon(applies), keyword)
	ab.Label = "Gives " + keyword + " as long as you've completed a dungeon."
	return ab
}

// Venturing is "venture into the dungeon" followed by the rest of the
// same instruction ("Venture into the dungeon and you gain 3 life"),
// as one Do: nothing the rest does reads the venture's answer.
func Venturing(then ...Applier) Effect {
	return Do(append([]Applier{VentureIntoTheDungeon{}}, then...)...)
}
