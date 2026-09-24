package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ao, the Dawn Sky — Legendary Creature — Dragon Spirit {3}{W}{W},
// 5/4:
//
//	"Flying, vigilance
//	 When Ao dies, choose one —
//	 • Look at the top seven cards of your library. Put any number of
//	   nonland permanent cards with total mana value 4 or less from
//	   among them onto the battlefield. Put the rest on the bottom of
//	   your library in a random order.
//	 • Put two +1/+1 counters on each permanent you control that's a
//	   creature or Vehicle."
//
// The card #998 was filed for, and the proof that the one-field gap
// it named was the whole gap. Every other piece was already written:
// WhenThisDies is the trigger, TriggeredAbility.Modes (#764) is the
// "choose one" — asked as the ability goes on the stack (CR 603.3c),
// so the controller picks a mode with the board as it stands after
// Ao died — LookAtTopOfLibraryForEffect is the look,
// PutFromLibraryOntoBattlefield is the put, PutRestOnBottomInRandomOrder
// is the rest, and b23TotalManaValueAtMost is the rule.
//
// What was missing was the WIRE between the last two. "Total mana
// value 4 or less" judges the SET the player picks, not each card in
// it, and the pick primitive built its choose_cards prompt without
// forwarding game.ChooseCardsPrompt.Validate. It forwards it now
// (put_from_library.go), so the rule is enforced where every other
// set rule is: once, inside checkChooseCardsPicksLocked, which the
// submit path and internal/legal's enumerator both go through.
//
// The first mode states the primitive directly rather than calling
// LookAtTopThenMayPutOntoBattlefield, because the sentence helper
// takes no set rule — see its doc.
//
// Flying and vigilance are printed keywords and come off the imported
// card.
func init() {
	Register(Spec{
		OracleID:     "697ac261-263b-4879-bf45-6d10d23312ae",
		Name:         "Ao, the Dawn Sky",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			aoTheDawnSkyDiesTrigger(),
		},
	})
}

// aoTheDawnSkyDiesTrigger is the dies trigger and its two bullets.
// Each bullet declares its own body (ModeDoing), so the engine runs
// the chosen one at resolution (CR 608.2c) and this file writes no
// switch. Neither bullet targets.
func aoTheDawnSkyDiesTrigger() game.TriggeredAbility {
	t := WhenThisDies("Ao, the Dawn Sky — choose one",
		func(g *game.Game, item *game.StackItem) error { return nil })
	t.Modes = ChooseOne(
		ModeDoing("Look at the top seven cards of your library. Put any number of nonland permanent cards with total mana value 4 or less from among them onto the battlefield. Put the rest on the bottom of your library in a random order.",
			nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				player := item.Controller
				return PutFromLibraryOntoBattlefield{
					Player:   player,
					Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 7),
					Match:    Nonland(),
					Max:      0,
					Optional: true,
					// The set rule, the one field #998 added. The
					// bounds cannot say it: "any number" is a
					// ceiling of seven and a floor of zero, and
					// four one-drops and one four-drop are both
					// five-card answers only one of which is legal.
					Validate: b23TotalManaValueAtMost(4),
					Label:    "Ao, the Dawn Sky — put any number of nonland permanent cards with total mana value 4 or less onto the battlefield",
					Then:     PutRestOnBottomInRandomOrder,
				}.Apply(ctx)
			}),
		ModeDoing("Put two +1/+1 counters on each permanent you control that's a creature or Vehicle.",
			nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return aoTwoCountersOnEachCreatureOrVehicle(ctx, item.Controller)
			}),
	)
	return t
}

// aoTwoCountersOnEachCreatureOrVehicle is the second bullet: two
// +1/+1 counters on every permanent the controller controls that is
// a creature or a Vehicle. A crewed Vehicle is both and still gets
// two counters, not four — the set is collected once, by instance.
//
// Ao itself is never among them: it is in a graveyard by the time
// its own dies trigger resolves.
func aoTwoCountersOnEachCreatureOrVehicle(ctx *Context, controller uuid.UUID) error {
	var ids []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller != controller {
			continue
		}
		if c.IsCreature() || c.HasSubtype("Vehicle") {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (AddCounter{Target: id, Kind: "+1/+1", N: 2}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
