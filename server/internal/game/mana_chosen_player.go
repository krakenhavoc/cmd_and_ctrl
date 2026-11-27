package game

import "github.com/google/uuid"

// mana_chosen_player.go — #6101: a mana ability whose mana goes to a
// player the activator chooses, in a colour THAT player chooses.
//
//	Spectral Searchlight  "{T}: Choose a player. That player adds one
//	                       mana of any color they choose."
//
// Choosing a player is not targeting (CR 115.10a), so the ability is a
// mana ability (CR 605.1a: "it doesn't require a target … it could add
// mana to a player's mana pool when it resolves") and it resolves the
// moment it is activated (CR 605.3b). It can be activated wherever a
// mana ability can (CR 605.3a), mid-payment included.
//
// The owner decision on #6101 (2026-11-27): the ability pauses until
// the chosen player has named the colour.
//
//   - The activator names the player WITH the activation
//     (ManaAbilityParams.ChosenPlayer, `chosen_player` on the wire):
//     any player still in the game, themselves included. None is the
//     activator, the answer every caller that predates the field means.
//   - The activator choosing themselves is Birds of Paradise: the
//     colour can be named up front (Colors) or is asked as an ordinary
//     mana_pick. Nothing here runs.
//   - Anyone else: the cost is paid, and the ability waits on ONE
//     PendingChoiceManaForPlayer asked of the chosen player
//     (queueManaForChosenPlayerLocked). Their answer goes through
//     ResolveManaChoice, which adds the mana to THEIR pool (CR 106.4,
//     "That player adds") and announces that the activator's ability
//     has resolved. Triggered mana abilities that watch the tap (CR
//     605.1b) see the activator as the one who tapped it.
//   - The pick blocks the table (choice_gate.go) as a mana_pick does:
//     the ability has not finished resolving. A chosen player who
//     leaves the game takes the question with them (CR 800.4a): the
//     mana would enter a pool that has left with its player. A chosen
//     seat that is disconnected or played by a bot answers through the
//     same paths every prompt does (the auto-answer timer and the bot
//     seat's choice moves).
//   - The auto-tapper never plans the ability (autotap.go): its
//     contract is "no further player decisions", and another seat's
//     answer is one.

// PendingChoiceManaForPlayer is the colour pick a "Choose a player.
// That player adds one mana of any color they choose." ability asks of
// the chosen player when that player is not the activator (#6101).
// Chooser is the chosen player, whose pool the mana enters; FromPlayer
// is the activator; Source is the permanent. ColorOptions are the
// printed colours, the chooser's commander identity first. Answered
// with `color`, like a mana_pick (ResolveManaChoice). An on-disk
// identity: never renamed, never reused.
const PendingChoiceManaForPlayer PendingChoiceKind = "mana_pick_for_player"

// isManaPickKind reports whether a choice kind is answered by
// ResolveManaChoice: a mana_pick or a mana_pick_for_player.
func isManaPickKind(k PendingChoiceKind) bool {
	return k == PendingChoiceMana || k == PendingChoiceManaForPlayer
}

// manaChosenPlayerLocked is the activation's check of
// ManaAbilityParams.ChosenPlayer, before anything is paid: refused on
// an ability that chooses no player, and naming a player still in the
// game otherwise. Nil is the activator. A colour named up front is
// refused beside another player: they choose it. Returns the chosen
// player.
//
// Caller must hold g.mu.
func (g *Game) manaChosenPlayerLocked(activator uuid.UUID, ab ManaAbilityShape, params ManaAbilityParams) (uuid.UUID, error) {
	if !ab.ChosenPlayerAdds {
		if params.ChosenPlayer != uuid.Nil {
			return uuid.Nil, ErrInvalidParam
		}
		return activator, nil
	}
	if params.ChosenPlayer == uuid.Nil || params.ChosenPlayer == activator {
		return activator, nil
	}
	if !g.playerInGameLocked(params.ChosenPlayer) {
		return uuid.Nil, ErrPlayerNotFound
	}
	if len(params.Colors) > 0 {
		return uuid.Nil, ErrInvalidParam
	}
	return params.ChosenPlayer, nil
}

// queueManaForChosenPlayerLocked asks `chosen` for the colour of the
// mana `ab` adds, after the activator's cost has been paid. The output
// is the printed "{W|U|B|R|G}": one pick. Reports whether a pick was
// queued; a slot list with nothing to offer, or a chosen player gone
// since the check, asks nothing and adds nothing.
//
// The choice carries everything the mint needs, as a mana_pick does
// (the restrictions, the riders, the source snapshot), because the
// answer comes after the source may have gone.
//
// Caller must hold g.mu (write).
func (g *Game) queueManaForChosenPlayerLocked(activator, chosen, cardID uuid.UUID, ab *ManaAbilityShape, slots []ProducedManaEntry, kinds ManaSourceKinds) bool {
	cp := g.playerByIDLocked(chosen)
	if cp == nil || cp.Eliminated {
		return false
	}
	var options []string
	for _, slot := range slots {
		options = append(options, slot.Options...)
	}
	options = manaPickOptionsFor(g, dedupeColors(options), cp, false)
	if len(options) == 0 {
		return false
	}
	choice := PendingChoice{
		Kind:             PendingChoiceManaForPlayer,
		Chooser:          chosen,
		FromPlayer:       activator,
		Count:            1,
		Source:           cardID,
		Reason:           ab.Label,
		ColorOptions:     options,
		ManaRestrictions: restrictionsFor(g, ab, activator, cardID),
		ManaSourceKinds:  kinds,
		ManaRiders:       spendRidersFor(g, ab, cardID),
		ManaTapped:       ab.TapCost,
		ManaLabel:        ab.Label,
		ManaAbilityRun:   uuid.NewString(),
	}
	return g.QueueChoiceForEffect(choice) != uuid.Nil
}

// dedupeColors keeps the first of each colour, in order.
func dedupeColors(in []string) []string {
	var out []string
	for _, c := range in {
		if !containsColor(out, c) {
			out = append(out, c)
		}
	}
	return out
}

// manaPickActor is the player who activated the ability a mana pick
// belongs to: the activator, FromPlayer, for a mana_pick_for_player,
// and the chooser for every other pick.
func manaPickActor(c *PendingChoice) uuid.UUID {
	if c.Kind == PendingChoiceManaForPlayer && c.FromPlayer != uuid.Nil {
		return c.FromPlayer
	}
	return c.Chooser
}
