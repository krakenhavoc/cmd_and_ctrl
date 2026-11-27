package effects

import (
	"slices"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b662_helpers.go — shared pieces of Invasion batch 662 (#5616).

// oldB662BasicLandTypes are the five basic land types (CR 205.3i).
var oldB662BasicLandTypes = []string{"Plains", "Island", "Swamp", "Mountain", "Forest"}

// OldB662BasicLandTypesOf is the basic land types a land has, read through the
// layers as the card the engine hands over.
func OldB662BasicLandTypesOf(c game.Card) []string {
	if !c.IsLand() {
		return nil
	}
	var out []string
	for _, t := range oldB662BasicLandTypes {
		if c.HasSubtype(t) {
			out = append(out, t)
		}
	}
	return out
}

// oldB662KeepsALandOfEachType is Global Ruin's set rule over `present`, the
// basic land types the choosing player's lands have: every type is covered by a
// kept land, and each kept land is chosen for a different type it has. One land
// with two types may be chosen for either (or for both, by being the only land
// kept), but a player never keeps a land chosen for nothing.
func oldB662KeepsALandOfEachType(present []string) func(picked []game.Card) bool {
	return func(picked []game.Card) bool {
		if len(picked) == 0 || len(picked) > len(present) {
			return false
		}
		types := make([][]string, len(picked))
		for i, c := range picked {
			types[i] = OldB662BasicLandTypesOf(c)
		}
		for _, t := range present {
			covered := false
			for _, ts := range types {
				if slices.Contains(ts, t) {
					covered = true
				}
			}
			if !covered {
				return false
			}
		}
		used := map[string]bool{}
		var assign func(i int) bool
		assign = func(i int) bool {
			if i == len(types) {
				return true
			}
			for _, t := range types[i] {
				if used[t] || !slices.Contains(present, t) {
					continue
				}
				used[t] = true
				if assign(i + 1) {
					return true
				}
				used[t] = false
			}
			return false
		}
		return assign(0)
	}
}

// oldB662Legs are the per-seat set rules, one for each non-empty set of the five
// basic land types a player's lands can have. The keys name the types and are
// on-disk identities: never renamed, never reused.
var oldB662Legs = func() (legs [1 << 5]game.ChooseCardsThen) {
	for mask := 1; mask < 1<<len(oldB662BasicLandTypes); mask++ {
		var present []string
		for i, name := range oldB662BasicLandTypes {
			if mask&(1<<i) != 0 {
				present = append(present, name)
			}
		}
		legs[mask] = game.RegisterOrderedPickLegRule(
			"choose-cards/s82-old-b662-global-ruin-keep-"+strings.ToLower(strings.Join(present, "-")),
			oldB662KeepsALandOfEachType(present))
	}
	return legs
}()

func oldB662Mask(types []string) int {
	mask := 0
	for i, name := range oldB662BasicLandTypes {
		if slices.Contains(types, name) {
			mask |= 1 << i
		}
	}
	return mask
}

// oldB662Lands is a seat's lands that an effect may sacrifice, read through the
// layers.
func oldB662Lands(g *game.Game, seat uuid.UUID) []game.Card {
	var out []game.Card
	for _, id := range g.EffectSacrificeCandidatesForEffect(seat, nil) {
		if c, ok := g.LookupCardForEffect(id); ok && c.IsLand() {
			out = append(out, c)
		}
	}
	return out
}

// oldB662Present is the basic land types some land in `cards` has.
func oldB662Present(cards []game.Card) []string {
	var present []string
	for _, t := range oldB662BasicLandTypes {
		for _, c := range cards {
			if slices.Contains(OldB662BasicLandTypesOf(c), t) {
				present = append(present, t)
				break
			}
		}
	}
	return present
}

// oldB662Choosable is the lands of a board that have a basic land type.
func oldB662Choosable(board []game.Card) []game.Card {
	var out []game.Card
	for _, c := range board {
		if len(OldB662BasicLandTypesOf(c)) > 0 {
			out = append(out, c)
		}
	}
	return out
}

// oldB662Forced reports that a board needs no choice: every land on it has a
// basic land type and the whole of it is the one keep-set.
func oldB662Forced(board []game.Card) bool {
	choosable := oldB662Choosable(board)
	if len(choosable) != len(board) {
		return false
	}
	return oldB662KeepsALandOfEachType(oldB662Present(choosable))(choosable)
}

// OldB662GlobalRuin is "Each player chooses from the lands they control a land
// of each basic land type, then sacrifices the rest." Each player in turn order
// picks the lands to keep (Cataclysmic Gearhulk's ordered pick, with the land
// rule), then every other land goes at once.
func OldB662GlobalRuin(g *game.Game, item *game.StackItem) error {
	var legs []game.ChooseCardsPrompt
	var seats []uuid.UUID
	for _, seat := range ApnapLivingSeats(g) {
		board := oldB662Lands(g, seat)
		if len(board) == 0 {
			continue
		}
		seats = append(seats, seat)
		choosable := oldB662Choosable(board)
		if oldB662Forced(board) || len(choosable) == 0 {
			continue
		}
		present := oldB662Present(choosable)
		ids := make([]uuid.UUID, len(choosable))
		for i, c := range choosable {
			ids[i] = c.InstanceID
		}
		legs = append(legs, game.ChooseCardsPrompt{
			Chooser:    seat,
			FromPlayer: seat,
			Source:     item.SourceCardID,
			Question:   "Global Ruin — choose from your lands a land of each basic land type to keep; the rest are sacrificed",
			Cards:      ids,
			Min:        1,
			Max:        len(present),
			Zone:       game.ZoneBattlefield,
			ThenKey:    oldB662Legs[oldB662Mask(present)],
		})
	}
	if len(seats) == 0 {
		return nil
	}
	if len(legs) == 0 {
		return oldB662SacrificeRest(g, item.SourceCardID, seats, nil)
	}
	return g.ChoosePermanentsInOrderKeyedForEffect(legs, oldB662KeepRest, item.SourceCardID, game.RunData{Carry: seats})
}

// oldB662KeepRest is "then sacrifices the rest", keyed. Carry is every seat
// with a land, in turn order.
var oldB662KeepRest = game.RegisterPromptRunThen("prompt-run/s82-old-b662-global-ruin-keep", func(g *game.Game, r game.RunDone) error {
	return oldB662SacrificeRest(g, r.Source, r.Carry, r.Picks().Cards())
})

// oldB662SacrificeRest sacrifices, at once, every land the seats control that is
// neither among `picked` nor on a board that needed no choice.
func oldB662SacrificeRest(g *game.Game, source uuid.UUID, seats, picked []uuid.UUID) error {
	kept := map[uuid.UUID]bool{}
	for _, id := range picked {
		kept[id] = true
	}
	var rest []uuid.UUID
	for _, seat := range seats {
		board := oldB662Lands(g, seat)
		forced := oldB662Forced(board)
		for _, c := range board {
			if kept[c.InstanceID] || (forced && len(OldB662BasicLandTypesOf(c)) > 0) {
				continue
			}
			rest = append(rest, c.InstanceID)
		}
	}
	if len(rest) == 0 {
		return nil
	}
	return g.SacrificeAllKeyedForEffect(source, rest, game.KeyedTail{})
}

// OldB662CoalitionVictory is "You win the game if you control a land of each
// basic land type and a creature of each color." Read as the spell resolves,
// through the layers; a creature with two colors counts for both.
func OldB662CoalitionVictory(g *game.Game, item *game.StackItem) error {
	types := map[string]bool{}
	colors := map[string]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != item.Controller {
			continue
		}
		for _, t := range OldB662BasicLandTypesOf(c) {
			types[t] = true
		}
		if c.IsCreature() {
			for _, col := range c.EffectiveColors() {
				colors[strings.ToUpper(col)] = true
			}
		}
	}
	for _, t := range oldB662BasicLandTypes {
		if !types[t] {
			return nil
		}
	}
	for _, col := range []string{"W", "U", "B", "R", "G"} {
		if !colors[col] {
			return nil
		}
	}
	return WinTheGame{Player: item.Controller}.Apply(NewContext(g, item))
}

// OldB662SwapTappedCreatures is "Simultaneously untap all tapped creatures and
// tap all untapped creatures" (Breaking Wave): both sets are read first, so a
// creature is moved once.
func OldB662SwapTappedCreatures(g *game.Game, item *game.StackItem) error {
	var tapped, untapped []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if !c.IsCreature() {
			continue
		}
		if c.Tapped {
			tapped = append(tapped, c.InstanceID)
		} else {
			untapped = append(untapped, c.InstanceID)
		}
	}
	ctx := NewContext(g, item)
	if err := UntapEach(ctx, tapped); err != nil {
		return err
	}
	for _, id := range untapped {
		if err := (TapTarget{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

// OldB662TapAllNonwhiteCreatures is "Tap all nonwhite creatures" (Blinding
// Light), whoever controls them, read through the layers as the spell resolves.
func OldB662TapAllNonwhiteCreatures(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && !c.Tapped && !slices.Contains(c.EffectiveColors(), "W") {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (TapTarget{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

// OldB662DeathOrGlory is "Separate all creature cards in your graveyard into two
// piles. Exile the pile of an opponent's choice and return the other to the
// battlefield." The caster separates (Split the Spoils' PileSplit, with the
// opponent chosen first as "an opponent" is the caster's choice) and the
// opponent takes one pile.
func OldB662DeathOrGlory(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ids := GraveyardCardIDs(ctx, item.Controller, func(c game.Card) bool { return c.IsCreature() })
	if len(ids) == 0 || len(ctx.Opponents()) == 0 {
		return nil
	}
	return ChoosePlayer{
		Among:    Opponents,
		Question: "Death or Glory — choose an opponent; they choose one of the two piles you separate",
		ThenKey:  oldB662DeathOrGloryChosen,
		Carry:    ids,
	}.Apply(ctx)
}

var oldB662DeathOrGloryChosen = ChoosePlayerThen("option-pick/s82-old-b662-death-or-glory-chosen", func(ctx *Context, chosen uuid.UUID, carry []uuid.UUID) error {
	if chosen == uuid.Nil || len(carry) == 0 {
		return nil
	}
	you := ctx.Controller()
	return PileSplit{
		Splitter:      you,
		Chooser:       chosen,
		Owner:         you,
		SplitQuestion: "Death or Glory — separate your creature cards into two piles",
		PickQuestion:  "Death or Glory — choose a pile: it is exiled, the other returns to the battlefield",
		Cards:         carry,
		ThenTail:      ctx.ItemTail(oldB662DeathOrGlorySettled),
	}.Apply(ctx)
})

// oldB662DeathOrGlorySettled exiles the pile the opponent took and returns the
// other, together, under the caster's control. A card that left the graveyard
// in the meantime is skipped.
var oldB662DeathOrGlorySettled = PileTail("pile/s82-old-b662-death-or-glory-settled", func(ctx *Context, taken, left []uuid.UUID, _ game.TailOutcome) error {
	var exile []uuid.UUID
	for _, id := range taken {
		if z := ctx.Game.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneGraveyard {
			exile = append(exile, id)
		}
	}
	if len(exile) > 0 {
		if err := ctx.Game.ExileCardsKeyedForEffect(exile, game.KeyedTail{}); err != nil {
			return err
		}
	}
	if len(left) == 0 {
		return nil
	}
	return ReturnFromGraveyardTogether{Targets: left, Controller: ctx.Controller()}.Apply(ctx)
})

// OldB662BarrinsSpite is "Choose two target creatures controlled by the same
// player. Their controller chooses and sacrifices one of them. Return the other
// to its owner's hand." With one target gone, the other is the one sacrificed.
func OldB662BarrinsSpite(_ *game.StackItem, ctx *Context) error {
	var ids []uuid.UUID
	for slot := 0; slot < 2; slot++ {
		if t, ok := ctx.ClauseTarget(slot); ok && t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	first, found := ctx.Game.LookupCardForEffect(ids[0])
	if !found {
		return nil
	}
	owner := first.Controller
	if len(ids) == 1 {
		return SacrificePermanent{Target: ids[0]}.Apply(ctx)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    owner,
		FromPlayer: owner,
		Source:     ctx.Source(),
		Question:   "Barrin's Spite — choose one of the two creatures to sacrifice; the other returns to its owner's hand",
		Cards:      ids,
		Min:        1,
		Max:        1,
		ThenKey:    oldB662BarrinsSpitePick,
		LiveZone:   game.ZoneBattlefield,
		Carry:      ids,
		Asked:      game.NewAskedItem(ctx.Item),
	})
	return nil
}

// oldB662BarrinsSpitePick sacrifices the chosen creature, then bounces the other.
var oldB662BarrinsSpitePick = game.RegisterChooseCardsThen("choose-cards/s82-old-b662-barrins-spite", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) != 1 || len(r.Carry) != 2 {
		return nil
	}
	ctx, _ := AskedContext(g, r.Asked, r.Chooser, r.Source, nil)
	other := r.Carry[0]
	if other == r.Picked[0] {
		other = r.Carry[1]
	}
	if err := (SacrificePermanent{Target: r.Picked[0]}).Apply(ctx); err != nil {
		return err
	}
	return BounceToHand{Target: other}.Apply(ctx)
})
