package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ascent of the Worthy — Enchantment — Saga {1}{W}{B}:
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I, II — Choose a creature you control. Until your next turn, all
//	 damage that would be dealt to creatures you control is dealt to that
//	 creature instead.
//	 III — Return target creature card from your graveyard to the
//	 battlefield with a flying counter on it. That creature is an Angel
//	 Warrior in addition to its other types."
//
// I and II (ADR 0108 §9, amendment 2026-10-16, #2056): the creature is
// chosen as the chapter resolves, not targeted, and with none to choose
// the chapter does nothing. The redirection protects the creatures you
// control as the damage would be dealt (CR 611.2c: one that enters later
// is covered, one you lose control of is not) and never you. Damage to
// the chosen creature itself is dealt to it, and once it has left the
// battlefield or stopped being a creature the redirection does nothing
// (CR 614.9). Chapters I and II make two effects: damage that meets both
// is ordered by you, and each applies once (CR 614.5, 616.1).
//
// III: the creature returns with its flying counter (CR 122.1b, an
// "enters with" counter), then is an Angel Warrior in addition to its
// other types for as long as it stays on the battlefield, a layer-4
// effect pinned to the object it is (CR 611.2a, 400.7), as Liliana,
// Death's Majesty's −3 is. The types are read through the chapter's text
// (CR 612.2).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "d748bad4-dd4c-4553-9fcc-e462260f6ff3",
		Name:         "Ascent of the Worthy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, "Ascent of the Worthy — I: damage to your creatures is dealt to the chosen creature", ascentOfTheWorthyRedirect),
			ChapterTrigger(2, "Ascent of the Worthy — II: damage to your creatures is dealt to the chosen creature", ascentOfTheWorthyRedirect),
			ChapterTriggerTargeting(3, "Ascent of the Worthy — III: return a creature with a flying counter",
				targetCreatureInYourGraveyard(), ascentOfTheWorthyReturn),
		},
	})
}

func ascentOfTheWorthyRedirect(g *game.Game, item *game.StackItem) error {
	return ChoosePermanents{
		Question:   "Ascent of the Worthy — choose a creature you control",
		Candidates: humiliateCreatures,
		Keyed:      ascentOfTheWorthyChosen,
	}.Apply(NewContext(g, item))
}

// ascentOfTheWorthyChosen is the chosen creature taking the redirected
// damage, keyed (#4290).
var ascentOfTheWorthyChosen = RegisterPermanentsPick("prompt-run/ascent-of-the-worthy-redirect", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		cards := picked.Cards()
		if len(cards) != 1 {
			return nil
		}
		return RedirectDamage{
			Protect:           ShieldCreaturesYouControl,
			To:                RedirectToObject(cards[0]),
			UntilYourNextTurn: true,
		}.Apply(ctx)
	},
})

func ascentOfTheWorthyReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ids := legalTargetCardIDs(ctx)
	if len(ids) == 0 {
		return nil
	}
	return g.ReturnFromGraveyardWithCountersKeyedForEffect(ids[0], item.Controller, false,
		map[string]int{game.CounterFlying: 1}, ctx.ItemTail(ascentOfTheWorthyType))
}

// ascentOfTheWorthyType is the entry's tail as a keyed tail (#4253 slice
// 6): it is an Angel Warrior in addition to its other types, the two
// types as the spell's text reads them, which the asking item that rides
// the tail gives back. An on-disk key.
var ascentOfTheWorthyType = EnteredTail("ascent-of-the-worthy/angel-warrior", func(ctx *Context, entered uuid.UUID, _ game.TailOutcome) error {
	g := ctx.Game
	// A graveyard return keeps the card's ID.
	if entered == uuid.Nil || !onBattlefield(g, entered) {
		return nil
	}
	return ScopedEffectFor{
		Target:   entered,
		Mods:     []game.Mod{game.AddSubtypesMod(ctx.TextSubtypes("Angel", "Warrior")...)},
		Duration: g.PinnedTo(game.IndefiniteDuration(), entered),
		Label:    "Ascent of the Worthy — it is an Angel Warrior in addition to its other types",
	}.Apply(ctx)
})
