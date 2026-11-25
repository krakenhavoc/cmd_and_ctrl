package game

import "github.com/google/uuid"

// power_gated_goad.go is a goad whose reach is decided by power (#6098):
// Baeloth Barrityl, Entertainer's "creatures your opponents control with
// power less than Baeloth Barrityl's power are goaded".
//
// A static goad is a layer-6 write of goad's two CR 701.15b requirements
// (GoadRequirements, effects.GoadAttached). That cannot work when the
// condition is power: layer 6 runs before layer 7 (CR 613.1f, 613.4) and
// the P/T counters are added after the pass, so a layer-6 AppliesTo would
// compare two unfinished powers. Goad is no characteristic: "attacks each
// combat if able" is an effect on the game rules, and CR 613.11 applies
// those "after all other continuous effects have been applied". So the
// static only records the gate in layer 6 (Characteristic.PowerGatedGoads,
// which goes with the static's own abilities, CR 613.1f), and this fold,
// run once every layer has run, writes the requirements for each creature
// whose finished power is below its source's. Power_floor.go's shape.
//
// From the requirements on, nothing is new: the attack judge, the legal
// enumerator, the bot, Goaded() and the view all read
// Characteristic.AttackRequirements. The fold writes no power and no
// type, so every other permanent's finished values are final whatever
// order it visits them in, and a pump, a counter or the source leaving
// bumps the layer version, so the next read sees the new answer.

// PowerGatedGoad is one "with power less than ~'s power are goaded"
// reaching a creature: the source permanent whose finished power is the
// bar (and whose name the refusal sentence gives), and the goading player.
type PowerGatedGoad struct {
	Source uuid.UUID `json:"source"`
	By     uuid.UUID `json:"by"`
}

// foldPowerGatedGoadsLocked turns each PowerGatedGoad whose source is on
// the battlefield, and whose bar the creature's finished power is below,
// into goad's CR 701.15b requirements. Powers are read as
// PowerForComparison reads them (layers and P/T counters, unclamped).
//
// Caller must hold g.mu in write mode (the layer pass).
func (g *Game) foldPowerGatedGoadsLocked() {
	if g.Battlefield == nil {
		return
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.effective == nil || len(c.effective.PowerGatedGoads) == 0 {
			continue
		}
		power, _ := ptWithCounters(*c.effective, c.Counters)
		for _, pg := range c.effective.PowerGatedGoads {
			src := findBattlefieldCard(g, pg.Source)
			if src == nil || src.effective == nil {
				continue
			}
			bar, _ := ptWithCounters(*src.effective, src.Counters)
			if power >= bar {
				continue
			}
			c.effective.AttackRequirements = append(c.effective.AttackRequirements, GoadRequirements(src, pg.By)...)
		}
	}
}
