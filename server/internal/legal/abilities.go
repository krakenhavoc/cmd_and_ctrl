package legal

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// activateParams is the activate_ability wire payload for a catalog
// ability (ability_index set). Free-form S13.1 announcements are not
// enumerated — nothing says what they do.
type activateParams struct {
	SourceCardID string       `json:"source_card_id"`
	AbilityIndex int          `json:"ability_index"`
	Targets      []targetWire `json:"targets,omitempty"`
	SacrificeIDs []string     `json:"sacrifice_ids,omitempty"`
	Strict       bool         `json:"strict,omitempty"`
	AutoTap      bool         `json:"auto_tap,omitempty"`
}

// activatedMoves enumerates catalog activated abilities on the
// seat's permanents. Requires priority (checked by the caller).
// Mirrors game.ActivateCatalogAbility's validation: split second,
// sorcery-speed flag, tap cost (untapped + not summoning sick for
// creatures), sacrifice costs payable, life cost payable, mana cost
// affordable, targets legal.
func (e *enumerator) activatedMoves() {
	g, p := e.g, e.p
	if g.SplitSecondActive {
		return
	}
	speed := sorcerySpeedOpen(g, e.seat)
	for i := range g.Battlefield.Cards {
		source := &g.Battlefield.Cards[i]
		if source.Controller != e.seat {
			continue
		}
		abilities := game.ActivatedAbilitiesForCard(*source)
		for idx, ab := range abilities {
			if (ab.SorcerySpeed || ab.Cost.Loyalty != nil) && !speed {
				continue
			}
			// CR 606: a loyalty ability needs a planeswalker, one
			// activation per turn, and enough counters to pay a −N.
			// Mirrors ActivateCatalogAbility so a policy never
			// proposes a move the engine will bounce.
			if ab.Cost.Loyalty != nil {
				if !source.IsPlaneswalker() || g.LoyaltyActivatedThisTurn[source.InstanceID] {
					continue
				}
				if n := *ab.Cost.Loyalty; n < 0 && source.Counters[game.CounterLoyalty] < -n {
					continue
				}
			}
			if ab.Cost.Tap {
				if source.Tapped {
					continue
				}
				if source.IsCreature() && game.HasSummoningSickness(source) {
					continue
				}
			}
			if ab.Cost.Life > 0 && p.Life < ab.Cost.Life {
				continue
			}
			if ab.Cost.Mana != "" {
				cost, err := game.ParseCost(ab.Cost.Mana)
				if err != nil || !e.canPay(cost, 0) {
					continue
				}
			}
			sacrificeSets := [][]uuid.UUID{nil}
			if ab.Cost.SacrificeOther != nil {
				pool := e.sacrificePool(source.InstanceID, ab.Cost.SacrificeSelf, ab.Cost.SacrificeOther)
				sacrificeSets = combinations(pool, 1, 1, e.opts.MaxExpansionPerSource)
				if len(sacrificeSets) == 0 {
					continue
				}
			}
			budget := e.opts.MaxExpansionPerSource
			targetSets := [][]game.TargetRef{nil}
			if ab.Targets != nil {
				targetSets = e.legalTargetSets(ab.Targets, budget)
				if len(targetSets) == 0 {
					continue
				}
			}
			for _, targets := range targetSets {
				for _, sacs := range sacrificeSets {
					if budget <= 0 {
						break
					}
					budget--
					label := source.Name + ": " + ab.Label + targetLabel(g, targets)
					e.add(Move{
						Type:   TypeActivateAbility,
						Player: e.seat,
						Kind:   KindActivate,
						Label:  label,
						Source: source.InstanceID,
						Params: mustJSON(activateParams{
							SourceCardID: source.InstanceID.String(),
							AbilityIndex: idx,
							Targets:      wireTargets(targets),
							SacrificeIDs: idStrings(sacs),
							Strict:       true,
							AutoTap:      true,
						}),
					})
				}
			}
		}
	}
}

// sacrificePool lists the seat's permanents that satisfy a
// SacrificeOther clause, excluding the source itself when the cost
// also sacrifices the source (the engine rejects paying one
// permanent twice).
func (e *enumerator) sacrificePool(sourceID uuid.UUID, selfToo bool, spec *game.TargetSpec) []uuid.UUID {
	// Sacrificing is a cost, not targeting, so this is the
	// candidate walk rather than the legal-TARGET walk — the
	// hexproof / shroud gate must not shrink the pool.
	lt := e.g.SpecCandidatesForEffect(e.seat, spec)
	var pool []uuid.UUID
	for _, id := range lt.Cards {
		if selfToo && id == sourceID {
			continue
		}
		if c := findBattlefield(e.g, id); c != nil && c.Controller == e.seat {
			pool = append(pool, id)
		}
	}
	return pool
}

type manaParams struct {
	CardID       string   `json:"card_id"`
	AbilityIndex int      `json:"ability_index"`
	SacrificeIDs []string `json:"sacrifice_ids,omitempty"`
}

// manaMoves enumerates mana abilities on the seat's permanents.
// Casts already auto-tap, so a policy rarely needs these for plain
// "{T}: Add" sources; they matter for sacrifice sources (Treasure,
// Lotus Petal, Ashnod's Altar) the auto-tapper never touches.
// Requires priority (checked by the caller). Mana abilities ignore
// split second (CR 702.61b).
func (e *enumerator) manaMoves() {
	g := e.g
	for i := range g.Battlefield.Cards {
		source := &g.Battlefield.Cards[i]
		if source.Controller != e.seat {
			continue
		}
		abilities := game.ManaAbilitiesForCard(*source)
		for idx, ab := range abilities {
			if ab.TapCost {
				if source.Tapped {
					continue
				}
				if source.IsCreature() && game.HasSummoningSickness(source) {
					continue
				}
			}
			sacrificeSets := [][]uuid.UUID{nil}
			if ab.SacrificeOther != nil {
				pool := e.sacrificePool(source.InstanceID, ab.SacrificeCost, ab.SacrificeOther)
				sacrificeSets = combinations(pool, 1, 1, e.opts.MaxExpansionPerSource)
				if len(sacrificeSets) == 0 {
					continue
				}
			}
			for _, sacs := range sacrificeSets {
				label := source.Name + ": " + ab.Label
				if ab.Label == "" {
					label = source.Name + ": add " + ab.Produced
				}
				if len(sacs) == 1 {
					label += " (sacrificing " + cardName(g, sacs[0]) + ")"
				}
				e.add(Move{
					Type:   TypeActivateManaAbility,
					Player: e.seat,
					Kind:   KindMana,
					Label:  label,
					Source: source.InstanceID,
					Params: mustJSON(manaParams{
						CardID:       source.InstanceID.String(),
						AbilityIndex: idx,
						SacrificeIDs: idStrings(sacs),
					}),
				})
			}
		}
	}
}
