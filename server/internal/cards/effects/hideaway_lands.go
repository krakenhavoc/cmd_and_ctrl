package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hideaway_lands.go — the Lorwyn-block hideaway lands (ADR 0091), the
// three that are on the card-coverage roadmap: Mosswort Bridge (batch
// 01, #294), Windbrisk Heights (batch 05, #298) and Spinerock Knoll
// (batch 06, #299).
//
// One shape, written as a table: "Hideaway 4. This land enters tapped.
// {T}: Add {C}. {C}, {T}: You may play the exiled card without paying
// its mana cost if <condition>." (CR 702.75b is why the Oracle text now
// prints "Hideaway 4" and the enters-tapped line — the errata that
// replaced the old fixed-four, enters-tapped keyword.)
//
// The condition is part of the EFFECT ("you may play … if …"), not an
// activation restriction, so it is asked when the ability RESOLVES: the
// land may be activated at any time, and pays for nothing if the
// condition is false then. The play itself is PlayHiddenCard over the
// card THIS land hid — read off the activation's object reference
// (HiddenRefOfActivation), so a land destroyed in response still plays
// its card, and a land that was bounced and replayed has hidden a new
// card and has no claim on the old one (CR 607.2a, CR 400.7).
//
// A hidden LAND is played during the resolution with the turn's land
// drop (CR 608.2g, CR 305.2a); a hidden SPELL is a {0} grant cast right
// after (ADR 0091, the cascade / Malcolm posture) — see PlayHiddenCard.

type hideawayLand struct {
	oracle string
	name   string
	color  string // the mana symbol, "W"
	// condition is the printed "if …", asked at resolution for the
	// ability's controller.
	condition func(g *game.Game, controller uuid.UUID) bool
	clause    string
}

var hideawayLands = []hideawayLand{
	{
		oracle:    "7cb9e29f-835f-4155-a2a5-4b778866c773",
		name:      "Mosswort Bridge",
		color:     "G",
		condition: creaturesYouControlTotalPowerAtLeast(10),
		clause:    "if creatures you control have total power 10 or greater",
	},
	{
		oracle:    "3589bcfc-42b0-414a-adce-bc690dc631c8",
		name:      "Windbrisk Heights",
		color:     "W",
		condition: attackedWithAtLeastThisTurn(3),
		clause:    "if you attacked with three or more creatures this turn",
	},
	{
		oracle:    "690c7f8e-fea2-4920-afa7-02ff120701a1",
		name:      "Spinerock Knoll",
		color:     "R",
		condition: anOpponentWasDealtAtLeastThisTurn(7),
		clause:    "if an opponent was dealt 7 or more damage this turn",
	},
}

func init() {
	for _, land := range hideawayLands {
		land := land
		mana := "{" + land.color + "}"
		label := mana + ", {T}: You may play the exiled card without paying its mana cost " + land.clause + "."
		Register(Spec{
			OracleID:     land.oracle,
			Name:         land.name,
			Completeness: CompletenessFull,
			Replacements: []game.ReplacementEffect{SelfEntersTapped()},
			Triggered:    []game.TriggeredAbility{Hideaway(land.name, 4)},
			ManaAbilities: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: mana,
				Label:    "Add " + mana,
			}},
			Activated: []ActivatedAbility{{
				Label: label,
				Cost:  Plus(ManaCost(mana), TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !land.condition(g, item.Controller) {
						return nil
					}
					return PlayHiddenCard{
						Source: HiddenRefOfActivation(item),
						Label:  land.name + " — play the exiled card without paying its mana cost",
					}.Apply(NewContext(g, item))
				},
			}},
		})
	}
}

// creaturesYouControlTotalPowerAtLeast is Mosswort Bridge's "creatures
// you control have total power N or greater" — current power, counters
// included (CurrentPower, #1281), post-layer.
func creaturesYouControlTotalPowerAtLeast(n int) func(*game.Game, uuid.UUID) bool {
	return func(g *game.Game, controller uuid.UUID) bool {
		total := 0
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == controller && c.IsCreature() {
				total += c.CurrentPower()
			}
		}
		return total >= n
	}
}

// attackedWithAtLeastThisTurn is Windbrisk Heights' "you attacked with
// N or more creatures this turn": N DISTINCT creatures `controller`
// declared as attackers this turn, across every combat — Minas
// Tirith's reading of the same words.
func attackedWithAtLeastThisTurn(n int) func(*game.Game, uuid.UUID) bool {
	return func(g *game.Game, controller uuid.UUID) bool {
		seen := map[uuid.UUID]bool{}
		for _, ev := range g.EventsThisTurn() {
			if !attackDeclaredByYou(ev, controller) || ev.CardID == uuid.Nil {
				continue
			}
			seen[ev.CardID] = true
			if len(seen) >= n {
				return true
			}
		}
		return false
	}
}

// anOpponentWasDealtAtLeastThisTurn is Spinerock Knoll's "an opponent
// was dealt N or more damage this turn": for some ONE opponent, the
// damage dealt to them this turn — combat or not, from any source —
// totals N. Damage dealt, not life lost: a Spinerock Knoll does not
// care about a drain, and a prevented point was never dealt (no
// EventDealDamage is emitted for it).
func anOpponentWasDealtAtLeastThisTurn(n int) func(*game.Game, uuid.UUID) bool {
	return func(g *game.Game, controller uuid.UUID) bool {
		dealt := map[uuid.UUID]int{}
		for _, ev := range g.EventsThisTurn() {
			if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Target == controller {
				continue
			}
			if g.PlayerByIDForEffect(ev.Target) == nil {
				continue
			}
			dealt[ev.Target] += ev.Amount
			if dealt[ev.Target] >= n {
				return true
			}
		}
		return false
	}
}
