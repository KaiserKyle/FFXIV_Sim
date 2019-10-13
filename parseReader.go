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
				skill := parsedSkill{sp[3], sp[7], sp[5], Time.Sub(p.startTime).Seconds() + 0.01, AbilityID, Flags, Damage, Crit, Direct, false}
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
