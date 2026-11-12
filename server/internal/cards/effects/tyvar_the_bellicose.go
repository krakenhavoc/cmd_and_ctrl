package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tyvar the Bellicose — Legendary Creature — Elf Warrior {2}{B}{G}, 5/4
// (March of the Machine: The Aftermath, #5644):
//
//	"Whenever one or more Elves you control attack, they gain deathtouch
//	 until end of turn.
//	 Each creature you control has \"Whenever a mana ability of this
//	 creature resolves, put a number of +1/+1 counters on it equal to the
//	 amount of mana this creature produced. This ability triggers only
//	 once each turn.\""
//
// The first ability is one trigger per attack declaration that includes
// an Elf you control (OncePerBatch, CR 603.2c), Tyvar among them. "They"
// are that declaration's attackers you control that are Elves, read from
// the declaration's own batch as the trigger resolves: a creature that
// only entered attacking was not declared (CR 508.4) and gets nothing.
// "Elf" is a word of the source's text and creature types are read
// through the layers, so a changeling counts.
//
// The second is a layer 6 grant of a trigger bundle (ADR 0093) to every
// creature you control, so it is each creature's own ability: its
// controller controls it, an ability removal on the creature takes it,
// and "only once each turn" is the engine's per-object trigger tally. It
// watches game.EventManaAbilityResolved (#5644), which every route
// through a mana ability emits once the ability has finished resolving:
// the hand click, the answered colour pick, the auto-tapper, and a
// triggered mana ability of the creature (CR 605.1b, Badgermole Cub's).
// The event's Amount is the mana that ability produced, after a
// CR 106.12b replacement (three from a Llanowar Elves under Nyxbloom
// Ancient) and without the mana another ability added because of it. The
// trigger goes on the stack the next time a player would receive
// priority (CR 603.3b), so the mana is already in the pool; it reads the
// amount its own event carried, and puts the counters on the creature
// only if it is still that permanent. An ability that resolved for no
// mana still triggers, uses up the turn and puts no counters.
//
// No simplification.
const (
	tyvarGrant      = "tyvar-the-bellicose/mana-counters"
	tyvarManaLabel  = "Tyvar the Bellicose — put +1/+1 counters on this creature equal to the mana it produced"
	tyvarTouchLabel = "Tyvar the Bellicose — those Elves gain deathtouch until end of turn"
)

func init() {
	Register(Spec{
		OracleID:     "a21d321b-2172-4577-a05b-399a019e3448",
		Name:         "Tyvar the Bellicose",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, tyvarAnElfYouControlAttacked, tyvarTouchLabel, tyvarDeathtouch)),
		},
		Grants: []AbilityGrant{{
			Key: tyvarGrant,
			Triggered: []game.TriggeredAbility{
				On(game.EventManaAbilityResolved,
					func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
						return source != nil && ev.CardID == source.InstanceID &&
							!b11TriggeredThisTurn(g, source.InstanceID, tyvarManaLabel)
					},
					tyvarManaLabel, tyvarCounters),
			},
			Text: "Whenever a mana ability of this creature resolves, put a number of +1/+1 counters on it equal to the amount of mana this creature produced. This ability triggers only once each turn.",
		}},
		Static: []game.StaticAbility{GrantAbilities(CreaturesYouControl, tyvarGrant)},
	})
}

// tyvarAnElfYouControlAttacked is "one or more Elves you control attack":
// the source's controller declared a creature that is an Elf as it
// attacks.
func tyvarAnElfYouControlAttacked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.HasSubtype(source.TextSubtype("Elf"))
}

// tyvarDeathtouch gives deathtouch until end of turn to every Elf of the
// batch's declaration that the trigger's controller attacked with.
func tyvarDeathtouch(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	batch := item.Trigger.Event.Batch
	elf := ctx.TextSubtype("Elf")
	seen := map[uuid.UUID]bool{}
	for _, ev := range g.Events {
		if ev.Kind != game.EventAttack || ev.Batch != batch || ev.Actor != item.Controller || ev.CardID == uuid.Nil || seen[ev.CardID] {
			continue
		}
		seen[ev.CardID] = true
		c, ok := g.LookupCardForEffect(ev.CardID)
		if !ok || !onBattlefield(g, ev.CardID) || !c.IsCreature() || !c.HasSubtype(elf) {
			continue
		}
		grant := GrantKeywordUntilEOT{Target: ev.CardID, Keywords: []string{"deathtouch"}, Label: "Tyvar the Bellicose — deathtouch"}
		if err := grant.Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// tyvarCounters is the granted ability's body: +1/+1 counters on the
// creature that has it, as many as the mana the resolved ability
// produced. A package-level func, so it captures nothing and resolves
// against whatever game an undo restored.
func tyvarCounters(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Event.Amount < 1 {
		return nil
	}
	return CounterOnThis{Kind: game.CounterPlusOne, N: item.Trigger.Event.Amount}.Apply(NewContext(g, item))
}
