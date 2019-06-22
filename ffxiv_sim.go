package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
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

func globalLogIntSlice(verb logVerbosity, logString string, intSlice []int) {
	str := fmt.Sprintf("%v", intSlice)
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

func loadTextFile(fileName string) *os.File {
	file, err := os.Open(fileName)
	if err != nil {
		globalLog(Important, "Text load error: "+err.Error())
	}
	return file
}

func main() {
	rngSource := rand.NewSource(time.Now().UnixNano())
	rng = rand.New(rngSource)

	verbFlag := flag.Int("verbosity", 0, "Log Level")
	numIterations := flag.Int("iterations", 1, "Number of time to run the rotation file")
	validateParse := flag.String("parse", "", "Parse to be validated against simulator")
	flag.Parse()

	if *verbFlag >= 2 {
		*verbFlag = 1
	}
	globalVerb = logVerbosity(*verbFlag)

	var player playerCharacter
	var skills []playerSkill
	var skillQueue []string
	var playerStance stance

	loadJSONFile("./kaiserkyle.json", &player)
	loadJSONFile("./effects.json", &effectLibrary)
	loadJSONFile("./"+player.Class+".json", &skills)
	loadJSONFile(player.RotationFile, &skillQueue)

	enemy := enemy{"Dummy A", 99999999, 0, nil, nil, 0.0}
	enemies = append(enemies, &enemy)

	player.Skills = skills
	players = append(players, &player)

	if player.Class == "Dragoon" {
		playerStance = new(dragoonStance)
	}
	player.PlayerStance = playerStance

	sim := simulator{players, enemies}

	if *validateParse != "" {
		fmt.Println("Reading Parse:", *validateParse)
		reader := parseReader{}
		reader.loadParseFile(*validateParse)
	} else {
		for i := 0; i < *numIterations; i++ {
			player.reset()
			enemy.reset()

			result := sim.RunSkillQueue(skillQueue)
			result.log()
		}
	}
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
