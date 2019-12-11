package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type parsedSkill struct {
	Name         string
	TargetName   string
	Ability      string
	RelativeTime float64
	AbilityID    int64
	Flags        int64
	Damage       int64
	Crit         bool
	Direct       bool
	Simulated    bool
	Validated    bool
}

type parseReader struct {
	parseFile       *os.File
	playerLines     []string
	startTime       time.Time
	skills          []parsedSkill
	TotalDamage     int64
	DoTTicks        int64
	AutoAttacks     int64
	CritCount       uint64
	DirectCount     uint64
	CritDirectCount uint64
}

type parseMismatch struct {
	simResult    *skillResult
	lowerBound   float64
	upperBound   float64
	parsedDamage int64
}

func (p *parseReader) loadParseFile(fileName string) {
	p.parseFile = loadTextFile(fileName)
	scanner := bufio.NewScanner(p.parseFile)
	p.playerLines = make([]string, 0)
	p.skills = make([]parsedSkill, 0)
	p.TotalDamage = 0
	p.DoTTicks = 0
	p.AutoAttacks = 0
	targetID := ""
	for scanner.Scan() {
		thisLine := scanner.Text()
		if strings.Contains(thisLine, players[0].Name) || strings.Contains(thisLine, "DoT") {
			sp := strings.Split(thisLine, "|")
			if sp[0] == "21" || sp[0] == "22" /*&& sp[3] != sp[7]*/ {
				// Uncomment for debugging
				//fmt.Println(thisLine)
				targetSelf := false
				if sp[3] != sp[7] {
					if targetID == "" {
						targetID = sp[6]
					} else {
						if sp[6] != targetID {
							continue
						}
					}
				} else {
					targetSelf = true
				}
				p.playerLines = append(p.playerLines, thisLine)
				AbilityID, _ := strconv.ParseInt(sp[4], 16, 0)
				Flags, _ := strconv.ParseInt(sp[8], 16, 0)
				Damage, _ := strconv.ParseInt(sp[9], 16, 0)
				if targetSelf {
					Damage = 0
				} else {
					Damage = Damage >> 16
				}
				Crit := (Flags & 0x100) != 0
				Direct := (Flags & 0x200) != 0
				Time, _ := time.Parse(time.RFC3339Nano, sp[1])
				if p.startTime.IsZero() {
					p.startTime = Time
				}

				// Add 0.01 to skill time because simulation starts at 0.01
				skill := parsedSkill{sp[3], sp[7], sp[5], Time.Sub(p.startTime).Seconds() + 0.01, AbilityID, Flags, Damage, Crit, Direct, false, false}
				skillFound := false
				for i := range players[0].Skills {
					if players[0].Skills[i].Name == skill.Ability {
						skillFound = true
						break
					}
				}
				if skillFound {
					p.skills = append(p.skills, skill)
					p.TotalDamage += Damage
				}
				if skill.Ability == "Attack" {
					p.AutoAttacks++
					p.TotalDamage += Damage
				}
				if Crit {
					p.CritCount++
				}
				if Direct {
					p.DirectCount++
				}
				if Crit && Direct {
					p.CritDirectCount++
				}
				skill.log()
			}
			if sp[0] == "24" {
				p.DoTTicks++
				Time, _ := time.Parse(time.RFC3339Nano, sp[1])
				Damage, _ := strconv.ParseInt(sp[6], 16, 0)
				p.TotalDamage += Damage
				globalLog(Basic, fmt.Sprintf("[%06.2f] [DoT Tick] : %d", Time.Sub(p.startTime).Seconds(), Damage))
			}
			// Need to handle below message as well
			/* sp[0] == "24" || sp[0] == "26" || sp[0] == "30"*/
		}
	}
}

// This function is for Validating that the simulator is capable of producing the given
// parse. It will simulate the parse multiple times and look for discrepancies.
func (p *parseReader) ValidateParse(s simulator) simResult {
	var totalResult simResult
	var parseErrors []parseMismatch
	totalResult.skillsPerformed = make([]int, len(s.Players[0].Skills))
	simErrors := 0
	simSkills := 0
	validSkills := 0
	invalidSkills := 0

	// Override logging to only Important, as we are looping
	oldGlobalLogLevel := globalVerb
	globalVerb = Important

	// Loop through all skills in the parse
	for simSkills < len(p.skills) {
		s.Players[0].reset()
		enemies[0].reset()
		totalResult.timeEllapsed = 0
		for i, skill := range p.skills {
			// Skip Auto Attacks
			// TODO: Should be able to validate auto attack damage, need to figure out timing issues.
			if skill.Ability == "Attack" {
				continue
			}

			skillPerformed := false
			for skillPerformed == false {
				if totalResult.timeEllapsed >= skill.RelativeTime {
					// Perform current action
					actionResult := s.Players[0].performAction(skill.Ability, enemies[0])
					if actionResult != nil {
						enemies[0].applyDamage(actionResult)
						//actionResult.Log()
						actionResult.ApplyBuffs(*s.Players[0])
						totalResult.totalDamageDone += actionResult.DamageDone
						totalResult.skillsPerformed[actionResult.SkillIndex]++
						totalResult.skillCount++
						totalResult.totalHits++
						if actionResult.DidCrit {
							totalResult.critCount++
						}
						if actionResult.DidDirect {
							totalResult.directCount++
						}
						if actionResult.DidCrit && actionResult.DidDirect {
							totalResult.critDirectCount++
						}
					} else {
						globalLog(Important, "SKILL UNAVAILABLE ERROR: "+skill.Ability)
						simErrors++
					}

					skillPerformed = true
					// Reset timing, as the logs may have latency in them.
					s.Players[0].GCDTimeRemaining = 0
					s.Players[0].IsGCDOnCooldown = false
					s.Players[0].AnimationLock = 0

					// Check Damage if we have the same Crit/Direct status as the parsed attack
					if actionResult != nil && skill.Crit == actionResult.DidCrit && skill.Direct == actionResult.DidDirect {
						damage := actionResult.BaseDamageDone
						if actionResult.DidCrit {
							damage *= actionResult.CritMultiplier
						}
						if actionResult.DidDirect {
							damage *= 1.25
						}
						// Calculate the acceptable range for the parsed value, based on our simulated value.
						lowerBound := float64(damage) * 0.95
						upperBound := float64(damage) * 1.05
						lowerBound *= actionResult.DamageBuff
						upperBound *= actionResult.DamageBuff
						if float64(skill.Damage) >= lowerBound && float64(skill.Damage) <= upperBound {
							if !skill.Simulated {
								p.skills[i].Validated = true
							}
							validSkills++
						} else {
							if !skill.Simulated || (skill.Simulated && skill.Validated) {
								var err parseMismatch
								err.simResult = actionResult
								err.parsedDamage = skill.Damage
								err.lowerBound = lowerBound
								err.upperBound = upperBound
								parseErrors = append(parseErrors, err)
								p.skills[i].Validated = false
							}
							invalidSkills++
						}
						p.skills[i].Simulated = true
					}
					// For buff-only skills, as long as they are performed on time, consider them validated
					if actionResult != nil && skill.Damage == 0 {
						p.skills[i].Simulated = true
						p.skills[i].Validated = true
					}
				}

				// Advance time to take care of Dots, Cooldowns, and Autos
				results := enemies[0].advanceTime(timeIncrement)
				if results != nil {
					totalResult.dotTicks += len(results)
					for i := range results {
						totalResult.totalDamageDone += results[i].DamageDone
					}
				}

				timeResult := s.Players[0].advanceTime(timeIncrement, enemies[0])

				if timeResult != nil {
					totalResult.totalDamageDone += timeResult.DamageDone
					if timeResult.SkillName == "Attack" {
						totalResult.autoAttacks++
					}
					// TODO Figure out proper logic here
					// Due to latency issues, Autos will always be off by ~0.01 or 2.
					//if skill.Ability == "Attack" {
					//	if s.Players[0].TotalTime > skill.RelativeTime+0.1 {
					//		globalLog(Important, "SKILL TIMING ERROR: "+skill.Ability)
					//	}
					//	skillPerformed = true
					//} else {
					//	globalLog(Important, "AUTO ATTACK TIMING")
					//}
					if results != nil {
						results = append(results, *timeResult)
					} else {
						results = []skillResult{*timeResult}
					}
				}
				for _, result := range results {
					enemies[0].applyDamage(&result)
					//result.Log()
					result.ApplyBuffs(*s.Players[0])
					totalResult.totalHits++
					if result.DidCrit {
						totalResult.critCount++
					}
					if result.DidDirect {
						totalResult.directCount++
					}
					if result.DidCrit && result.DidDirect {
						totalResult.critDirectCount++
					}
				}
				totalResult.timeEllapsed += timeIncrement
			}
		}
		simSkills = 0
		validSkills = 0
		invalidSkills = 0
		for _, skill := range p.skills {
			if skill.Simulated {
				simSkills++
				if skill.Validated {
					validSkills++
				} else {
					invalidSkills++
				}
			}
		}
	}

	globalLog(Important, "Total Timing Errors: "+strconv.Itoa(simErrors))
	globalLog(Important, "Validated Skills: "+strconv.Itoa(validSkills))
	globalLog(Important, "Invalid Skills: "+strconv.Itoa(len(parseErrors)))

	globalVerb = oldGlobalLogLevel

	for _, err := range parseErrors {
		globalLog(Important, "Invalid Skill Result:")
		err.simResult.Log()
		globalLogFloat(Important, "Sim Upper Bound: ", err.upperBound)
		globalLogFloat(Important, "Sim Lower Bound: ", err.lowerBound)
		globalLog(Important, "Parsed Damage: "+strconv.FormatInt(err.parsedDamage, 10))
		globalLog(Important, "")
	}

	return totalResult
}

func (p *parsedSkill) log() {
	modifiers := ""
	if p.Crit {
		modifiers += "[CRIT]"
	}
	if p.Direct {
		modifiers += "[DIRECT]"
	}
	globalLog(Basic, fmt.Sprintf("[%06.2f] [%s -> %s] %s : %d %s", p.RelativeTime, p.Name, p.TargetName, p.Ability, p.Damage, modifiers))
}

func (p *parseReader) log() {
	globalLogFloat(Important, "Total Parsed Damage Done: ", float64(p.TotalDamage))
	globalLog(Important, "Skills performed: "+strconv.Itoa(len(p.skills)))
	globalLogFloat(Important, "Crit Rate: ", float64(p.CritCount)/float64(len(p.skills)))
	globalLogFloat(Important, "Direct Rate: ", float64(p.DirectCount)/float64(len(p.skills)))
	globalLogFloat(Important, "CritDirect Rate: ", float64(p.CritDirectCount)/float64(len(p.skills)))
	globalLogFloat(Important, "Auto Attacks: ", float64(p.AutoAttacks))
	globalLogFloat(Important, "DoT Ticks: ", float64(p.DoTTicks))
	globalLog(Important, "")
}
