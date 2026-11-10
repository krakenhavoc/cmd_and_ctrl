package game

import "github.com/google/uuid"

// emblem_tribe.go — an emblem that remembers a chosen creature type
// ("Choose a creature type. You get an emblem with 'Creatures you control
// of the chosen type get +3/+3 and have vigilance and hexproof.'", Oko,
// Shadowmoor Scion; CR 114, CR 205.3m).
//
// The emblem's statics already read the SOURCE's NamedTribe when the
// tribal filter is Chosen (effects.TribeFilter), and an emblem is the
// source of its own statics (CR 114.3), so the whole seam is stamping the
// choice on the emblem object as it is created. NamedTribe is an existing
// Card field, per instance and carried by the snapshot, so nothing about
// the snapshot's shape changes (#522).

// CreateEmblemWithTribeForEffect is CreateEmblemForEffect plus the
// creature type the effect's resolution chose, stored on the new emblem.
// An empty or non-canonical type creates no emblem and returns
// ErrInvalidParam: an effect with no choice made (the chooser left the
// game) makes nothing.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) CreateEmblemWithTribeForEffect(owner, sourceCardID uuid.UUID, tribe string) error {
	canonical, ok := CanonicalCreatureType(tribe)
	if !ok {
		return ErrInvalidParam
	}
	if err := g.CreateEmblemForEffect(owner, sourceCardID); err != nil {
		return err
	}
	p := g.playerByIDLocked(owner)
	if p == nil || p.Emblems == nil || len(p.Emblems.Cards) == 0 {
		return ErrPlayerNotFound
	}
	p.Emblems.Cards[len(p.Emblems.Cards)-1].NamedTribe = canonical
	g.layerVersion.Add(1)
	g.recomputeLayersLocked()
	return nil
}
