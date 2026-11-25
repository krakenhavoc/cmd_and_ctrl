package game

import "github.com/google/uuid"

// mana_color_of_choice.go — #6210: a mana-production replacement that
// asks for a colour. Pulse of Llanowar: "If a basic land you control is
// tapped for mana, it produces mana of a color of your choice instead of
// any other type." Harvest Mage: "… it produces one mana of a color of
// your choice instead of any other type and amount."
//
// CR 106.12b: "A replacement effect that applies if a permanent 'is
// tapped for mana' … modifies the mana production event while such an
// ability is resolving and producing mana." CR 605.3b: the ability
// resolves immediately, with no point inside it where a prompt could
// open, and RepEventProduceMana can never pause (produce_mana.go). So
// the colour is chosen AS PART OF THE TAP, not by a paused replacement
// (owner decision on #6210, 2026-11-24):
//
//   - The replacement writes ManaOfChosenColor for each mana it makes
//     ("of a color of your choice"), and the production resolves the
//     marker to the colour the tap chose.
//   - A tap by hand: the activation that the window would turn into that
//     marker becomes a colour pick over the five colours, the same pick
//     a Birds of Paradise asks (activateManaAbilityLocked), or the colour
//     the activator named up front. The pick's colour is what the
//     production asks for, and the marker resolves to it.
//   - The auto-tapper: the planner prices the land as a five-colour
//     source (priceProducedSlotsLocked), and the executor's production
//     resolves the marker to the colour its ledger still needs.

// ManaOfChosenColor is the mana a production replacement writes for "a
// mana of a color of your choice": one entry per mana, resolved to a
// colour before anything reaches the pool. Never in a pool, an event or
// a snapshot.
const ManaOfChosenColor = "*chosen*"

// allManaOfChosenColor reports whether a previewed production is
// entirely chosen-colour mana.
func allManaOfChosenColor(made []string) bool {
	if len(made) == 0 {
		return false
	}
	for _, c := range made {
		if c != ManaOfChosenColor {
			return false
		}
	}
	return true
}

// resolveChosenManaColors replaces each ManaOfChosenColor in `produced`
// with the tap's colour: the colour the auto-tapper's ledger needs most
// when there is one (pending, not booked here: bookProduced books what
// arrived), and otherwise the colour the production asked for, which a
// pick by hand has already made the chosen one. A production that asked
// for colourless mana and has no need to steer it picks white, the first
// colour, so no marker ever reaches the pool.
func resolveChosenManaColors(asked, produced []string, pending *[]ColorRequirement) []string {
	var choice string
	for i, c := range produced {
		if c != ManaOfChosenColor {
			continue
		}
		if choice == "" {
			if pending != nil {
				if best, opt := mostRestrictiveRequirement(AllColors, *pending); best >= 0 {
					choice = opt
				}
			}
			if choice == "" && len(asked) > 0 {
				a := asked[min(i, len(asked)-1)]
				if containsColor(AllColors, a) {
					choice = a
				}
			}
			if choice == "" {
				choice = AllColors[0]
			}
		}
		produced[i] = choice
	}
	return produced
}

// producesChosenColorLocked reports whether tapping cardID for `colors`
// would come out of the window as mana of a colour of the player's
// choice, and how much of it (Pulse keeps the amount, Harvest Mage makes
// one). Read-only.
//
// Caller must hold g.mu.
func (g *Game) producesChosenColorLocked(playerID, cardID uuid.UUID, colors []string, fromTap bool) (int, bool) {
	made := g.producedManaPreviewLocked(playerID, cardID, colors, fromTap)
	if !allManaOfChosenColor(made) {
		return 0, false
	}
	return len(made), true
}

// ModLandManaOfChosenColor is Harvest Mage's "Until end of turn, if you
// tap a land for mana, it produces one mana of a color of your choice
// instead of any other type and amount" (#6210). Reads Player, the
// "you" whose taps it replaces. Scope ScopeGame. A replacement kind
// (readerReplacement), never the layer pass. An on-disk identity: never
// renamed, never reused.
const ModLandManaOfChosenColor ModKind = "landManaOfChosenColor"

// LandsYouTapProduceOneManaOfAColorOfYourChoiceThisTurnForEffect
// registers Harvest Mage's replacement until end of turn for `player`.
// Reports whether a record was written.
//
// Caller must hold g.mu (write).
func (g *Game) LandsYouTapProduceOneManaOfAColorOfYourChoiceThisTurnForEffect(sourceID, player uuid.UUID, label string) bool {
	if g.playerByIDLocked(player) == nil {
		return false
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, player,
		[]Mod{{Kind: ModLandManaOfChosenColor, Player: player}}, g.UntilEndOfTurnDuration(), label)
}

// landManaOfChosenColorAppliesLocked: a land `m.Player` taps for mana
// (CR 106.12), making any mana at all. "Instead of any other type and
// amount", so colourless mana and a production already of one mana are
// replaced too.
func landManaOfChosenColorAppliesLocked(g *Game, m Mod, ev *ReplacementEvent) bool {
	if ev.Kind != RepEventProduceMana || !ev.ManaFromTap || len(ev.ManaColors) == 0 || ev.ManaPlayer != m.Player {
		return false
	}
	land, ok := g.LookupCardForEffect(ev.ManaSource)
	return ok && land.IsLand()
}

// steerChosenColorLocked is the auto-tap executor's half: a source whose
// production the window turns into chosen-colour mana makes the colour
// the cast's ledger still needs most, so the executor asks for that
// colour (and books it) instead of the printed one. `asked` and `booked`
// come back unchanged for every other source.
//
// Caller must hold g.mu.
func (g *Game) steerChosenColorLocked(playerID, cardID uuid.UUID, asked, booked []string, fromTap bool, pending []ColorRequirement) ([]string, []string) {
	if len(asked) == 0 {
		return asked, booked
	}
	if _, ok := g.producesChosenColorLocked(playerID, cardID, asked, fromTap); !ok {
		return asked, booked
	}
	best, color := mostRestrictiveRequirement(AllColors, pending)
	if best < 0 {
		return asked, booked
	}
	return repeatColor(color, len(asked)), repeatColor(color, len(booked))
}

// widenChosenColorSlotsLocked is a mana ability's slots as a player
// naming colours up front sees them (mana_color_upfront.go): a slot the
// window turns into chosen-colour mana offers the five colours, so the
// client can publish them and the activator can name one.
//
// Caller must hold g.mu.
func (g *Game) widenChosenColorSlotsLocked(playerID, cardID uuid.UUID, slots []ProducedManaEntry, fromTap bool) []ProducedManaEntry {
	var out []ProducedManaEntry
	for i, slot := range slots {
		if len(slot.Options) == 0 || slot.DifferentColors() {
			continue
		}
		first := slot.Options[0]
		if _, ok := g.producesChosenColorLocked(playerID, cardID, repeatColor(first, slot.AmountFor(first)), fromTap); !ok {
			continue
		}
		if out == nil {
			out = append([]ProducedManaEntry(nil), slots...)
		}
		out[i] = ProducedManaEntry{Options: append([]string(nil), AllColors...)}
	}
	if out == nil {
		return slots
	}
	return out
}
