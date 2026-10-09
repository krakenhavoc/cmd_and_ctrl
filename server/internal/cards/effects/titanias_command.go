package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Titania's Command — Sorcery {4}{G}{G} (EDHREC rank 4006):
//
//	"Choose two —
//	 • Exile target player's graveyard. You gain 1 life for each card
//	   exiled this way.
//	 • Search your library for up to two land cards, put them onto
//	   the battlefield tapped, then shuffle.
//	 • Create two 2/2 green Bear creature tokens.
//	 • Put two +1/+1 counters on each creature you control."
//
// The green Command, and six mana for two of four real effects is why
// it sees play over any one of them: ramp plus bodies, or graveyard
// hate plus a team pump, chosen after you know what the table is
// doing.
//
// It is in the batch as the four-option "choose two" whose FIRST
// option targets and whose other three do not — the shape the modal
// clause is limited to today (at most one chosen option may carry a
// target slot). Austere Command is the untargeted version of the same
// structure; this is the version with a target in it.
//
// The modes resolve in PRINTED ORDER (CR 608.2c) regardless of which
// order they were picked in, which matters for two pairings here:
//
//   - bullets three and four together make the Bears first and then
//     put counters on them, so two 2/2s become two 4/4s;
//   - bullets two and four together finish the search before the
//     counters: a creature land the search finds is on the battlefield
//     when the counters go out, so it gets them too. The search is a
//     prompt, and each bullet is a ModeDoing body so the engine's
//     bullet walk holds the later ones until it is answered (#2789).
//     As one OnResolve the counters ran on the line below the search,
//     while the prompt was still open.
//
// Notes on the individual bullets:
//
//   - The exile counts what actually REACHED exile and gains that
//     much life, so a card removed in response pays nothing. "Target
//     player" includes yourself.
//   - "Up to two land cards" may find one or none, and the lands come
//     in tapped. Any land card, not just basics.
//   - The counters go on every creature its controller controls at
//     resolution, including the Bears and any creature land from the
//     bullets above.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7aae0a3d-8882-4485-a126-f06ea6593dcb",
		Name:         "Titania's Command",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose two", 2, 2,
			ModeDoing("Exile target player's graveyard. You gain 1 life for each card exiled this way.",
				TargetPlayer("target player"), titaniasCommandExileAndGain),
			ModeWithPurpose(ModeDoing("Search your library for up to two land cards, put them onto the battlefield tapped, then shuffle.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return SearchLibrary{
						Player:        item.Controller,
						Predicate:     game.Card.IsLand,
						Dest:          game.ZoneBattlefield,
						Limit:         2,
						TappedOnEntry: true,
						Shuffle:       true,
						Optional:      true,
						Reason:        "Titania's Command — up to two land cards, onto the battlefield tapped",
					}.Apply(ctx)
				}), game.Purpose{Lands: 2}),
			ModeWithPurpose(ModeDoing("Create two 2/2 green Bear creature tokens.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return CreateToken{
						Controller: item.Controller,
						Template:   TokenCard("2/2 green Bear"),
						N:          2,
					}.Apply(ctx)
				}), game.Purpose{Tokens: 2}),
			ModeDoing("Put two +1/+1 counters on each creature you control.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return titaniasCommandCountersOnYourCreatures(ctx)
				}),
		),
	})
}

// titaniasCommandExileAndGain is the first bullet: exile a player's
// graveyard, then gain 1 life per card that actually reached exile.
//
// The count is taken AFTER the move, for the reason Canoptek Scarab
// Swarm's is: "each card exiled this way" is the cards that got
// there, not the cards that were in the pile when the mode was
// chosen.
func titaniasCommandExileAndGain(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	p := ctx.Game.PlayerByIDForEffect(t.ID)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(p.Graveyard.Cards))
	for _, c := range p.Graveyard.Cards {
		ids = append(ids, c.InstanceID)
	}
	exiled := 0
	for _, id := range ids {
		if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
		if z := ctx.Game.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
			exiled++
		}
	}
	return GainLife{Player: ctx.Controller(), Amount: exiled}.Apply(ctx)
}

// titaniasCommandCountersOnYourCreatures is the fourth bullet: two
// +1/+1 counters on each creature its controller controls, snapshotted
// before the first counter lands so a creature something makes on the
// way (a Doubling Season token) is not caught mid-loop.
func titaniasCommandCountersOnYourCreatures(ctx *Context) error {
	var yours []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.Controller == ctx.Controller() {
			yours = append(yours, c.InstanceID)
		}
	}
	for _, id := range yours {
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 2}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
