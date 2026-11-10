package game

import "github.com/google/uuid"

// mana_source_color.go — "Add one mana of any color. This creature
// becomes that color until end of turn." (Foraging Wickermaw), the
// engine half of ManaAbilityShape.SourceBecomesProducedColor.
//
// The colour is the one the activation produced, so it is applied where
// the colours become known: at once when they are settled as the
// ability resolves (CR 605.3b), or when the controller answers the
// colour pick (ResolveManaChoice, carried there as
// PendingChoice.ManaRecolorsSource). The effect is a layer 5
// colour-setting effect (CR 613.1e) pinned to the permanent as it is
// now, so a source that has left the battlefield gets nothing and a new
// object the card becomes is not recoloured (CR 400.7).

// sourceBecomesManaColorLocked makes the permanent `source` the
// colour(s) among `produced` until end of turn. Colourless mana names
// no colour, and an activation that produced none changes nothing.
//
// Caller must hold g.mu (write).
func (g *Game) sourceBecomesManaColorLocked(source uuid.UUID, produced []string) {
	var colors []string
	for _, c := range produced {
		switch c {
		case "W", "U", "B", "R", "G":
			if !typeListHas(colors, c) {
				colors = append(colors, c)
			}
		}
	}
	if len(colors) == 0 {
		return
	}
	live, ok := g.battlefieldCardLocked(source)
	if !ok {
		return
	}
	g.RegisterScopedEffectForEffect(source,
		[]AffectedObject{PinObject(live.InstanceID, live.EnteredBattlefieldAt)},
		[]Mod{SetColorsMod(colors...)}, g.UntilEndOfTurnDuration(),
		live.Name+" — becomes that color until end of turn")
}
