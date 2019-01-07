package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"time"
)

var rng *rand.Rand
var effectLibrary []effect
var players []*playerCharacter
var enemies []*enemy

type logVerbosity int

const (
	Basic = iota
	Important
)

var globalVerb logVerbosity

type buffIndex int

const (
	DamageDealt = iota
	CritRate
)

type buffAppliesTo int

const (
	All = iota
	Weaponskill
)

type enemyDebuffIndex int

const (
	EnemyVulnUp = iota
	EnemyPiercing
)

// No pointers, or else applyAdditionalEffect will need to do a deep copy.
// Duration is stored in seconds
type effect struct {
	Name                string
	Target              string
	Duration            int
	OffensiveBuffs      []int
	DefensiveBuffs      []int
	TimeLeft            float64
	AppliesTo           buffAppliesTo
	OnlyNextWeaponskill bool
	DoTPotency          int
}

func globalLog(verb logVerbosity, logString string) {
	if verb >= globalVerb {
		fmt.Println(logString)
	}
}

func globalLogFloat(verb logVerbosity, logString string, floatVal float64) {
	str := fmt.Sprintf("%f", floatVal)
	globalLog(verb, logString+str)
}

func loadJSONFile(fileName string, obj interface{}) {
	raw, err := ioutil.ReadFile(fileName)
	if err != nil {
		globalLog(Important, "JSON load error: "+err.Error())
	}

	err = json.Unmarshal(raw, obj)
	if err != nil {
		globalLog(Important, "Unmarshal error: "+err.Error())
	}
}

func main() {
	rngSource := rand.NewSource(time.Now().UnixNano())
	rng = rand.New(rngSource)

	verbFlag := flag.Int("verbosity", 0, "Log Level")
	flag.Parse()

	globalVerb = logVerbosity(*verbFlag)

	var player playerCharacter
	var skills []playerSkill
	var skillQueue []string
	var playerStance stance

	loadJSONFile("./kaiserkyle.json", &player)
	loadJSONFile("./effects.json", &effectLibrary)
	loadJSONFile("./"+player.Class+".json", &skills)
	loadJSONFile(player.RotationFile, &skillQueue)

	if player.Class == "Dragoon" {
		playerStance = new(dragoonStance)
	}
	player.Skills = skills
	player.PlayerStance = playerStance
	players = append(players, &player)

	enemy := enemy{"Dummy A", 99999999, 0, nil, nil, 0.0}
	enemies = append(enemies, &enemy)

	sim := simulator{players, enemies, skillQueue}
	sim.Run()
}

// global function to apply additional effects
func applyAdditionalEffect(effectName string, playerData playerCharacter, enemyName string) {
	if effectName == "" {
		return
	}
	for i := range effectLibrary {
		if effectLibrary[i].Name == effectName {
			newEffect := effectLibrary[i]
			newEffect.TimeLeft = float64(newEffect.Duration)

			if effectLibrary[i].Target == "Self" {
				for i := range players {
					if players[i].Name == playerData.Name {
						players[i].applyEffect(newEffect)
					}
				}
			} else if effectLibrary[i].Target == "Area" {
				// INCOMPLETE, must check for players in range
				for i := range players {
					if players[i].Name == playerData.Name {
						players[i].applyEffect(newEffect)
					}
				}
			} else if effectLibrary[i].Target == "Enemy" {
				for i := range enemies {
					if enemies[i].Name == enemyName {
						if newEffect.DoTPotency != 0 {
							// Create DoT
							result := new(skillResult)

							result.SkillName = effectName
							result.PlayerName = playerData.Name
							result.TargetName = enemyName

							result.calculateDamage(nil, playerData, false, newEffect.DoTPotency)
							enemies[i].applyDamage(result)

							dotResult := new(dotEffect)
							dotResult.Name = effectName
							dotResult.PlayerName = playerData.Name
							dotResult.Duration = float64(newEffect.Duration)
							dotResult.BaseDamage = int(math.Floor(result.BaseDamageDone * result.DamageBuff))
							dotResult.CritChance = result.CritRate
							dotResult.DirectChance = result.DirectRate
							dotResult.CritDamage = int(math.Floor(float64(dotResult.BaseDamage) * result.CritMultiplier))

							enemies[i].applyDoTEffect(*dotResult)
						} else {
							enemies[i].applyEffect(newEffect)
						}
					}
				}
			}
			break
		}
	}
}
