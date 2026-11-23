package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exhaust_cards_test.go — #1181, the catalog half of the exhaust
// keyword. The engine half (the record, its two scopes, the CR 400.7
// key, the clone and the snapshot) is in game/exhaust_test.go; the
// enumerator's and the wire's halves are in internal/legal and
// internal/protocol. Everything here is a real printed card.
//
// Four proofs, chosen for what each one adds:
//
//	Prowcatcher Specialist  the keyword and nothing else
//	Greenbelt Guardian      an exhaust ability BESIDE a repeatable one
//	Bitter Work             exhaust AND a printed activation condition,
//	                        over earthbend (#1180)
//	Ba Sing Se              an earthbend activation that is NOT exhaust

const (
	ProwcatcherSpecialistOracle = "594132ac-32a7-41d5-b7f0-3692bbed7f6f"
	GreenbeltGuardianOracle     = "2d8aa053-289d-40d9-baa7-9bd1c5b8e957"
	BitterWorkOracle            = "e9a24ed8-e844-4c50-b386-bc28da9989b1"
	BaSingSeOracle              = "de1ae205-ca5b-4d26-8194-ca85f1406e53"
)

// fillPoolColored drops n mana of one colour into a seat's pool.
// fillPool's colourless mana cannot pay a {R} or a {G}.
func FillPoolColored(p *game.Player, color string, n int) {
	for i := 0; i < n; i++ {
		p.ManaPool.AddMana(game.ManaToken{Color: color})
	}
}
