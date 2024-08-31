package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"

	"github.com/khiemnguyen15/discord-currency-converter/internal/conversions"
	"github.com/khiemnguyen15/discord-currency-converter/internal/currencyformatter"
)

var (
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
	WarningLogger *log.Logger
	DebugLogger   *log.Logger

	session *discordgo.Session
)

func init() {
	logFile, err := os.OpenFile(
		fmt.Sprintf("logs/logfile_%s.log", time.Now().Format("2006-01-02")),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0600,
	)
	if err != nil {
		log.Fatalln(err)
	}

	InfoLogger = log.New(logFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(logFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(logFile, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	DebugLogger = log.New(logFile, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)

	InfoLogger.Println("Process starting...")

	viper.SetConfigFile("config.yaml")
	err = viper.ReadInConfig()
	if err != nil {
		ErrorLogger.Fatalln("Error while reading config file: ", err)
	}

	currencyformatter.LoadTagMap()

	session, err = discordgo.New("Bot " + viper.GetString("discord.bot_token"))
	if err != nil {
		ErrorLogger.Fatalln("Invalid bot parameters: ", err)
	}
}

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "convert",
			Description: "Convert between two currencies",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "from",
					Description: "The base currency to convert from",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "to",
					Description: "The ending currency to convert to",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "value",
					Description: "The amount you want to convert",
					Required:    true,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"convert": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			optionMap := make(
				map[string]*discordgo.ApplicationCommandInteractionDataOption,
				len(options),
			)
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}

			fromOption, _ := optionMap["from"]
			toOption, _ := optionMap["to"]
			valueOption, _ := optionMap["value"]

			from := strings.ToUpper(fromOption.StringValue())
			to := strings.ToUpper(toOption.StringValue())
			value := valueOption.FloatValue()

			if value <= 0 {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Invalid value.",
					},
				})
				return
			}

			convertedValue, err := conversions.ConvertCurrency(from, to, value)
			if err != nil {
				ErrorLogger.Println(err)
				return
			}

			if convertedValue == 0 {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Invalid currency.",
					},
				})
				return
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title: fmt.Sprintf("%s to %s is:",
								currencyformatter.FormatCurrency(from, value),
								to,
							),
							Fields: []*discordgo.MessageEmbedField{
								{
									Value: currencyformatter.FormatCurrency(to, convertedValue),
								},
							},
							Color: 0xDDBD46,
						},
					},
				},
			})
		},
	}
)

func init() {
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func main() {
	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		InfoLogger.Printf(
			"Logged in as: %v#%v\n",
			s.State.User.Username,
			s.State.User.Discriminator,
		)
	})
	err := session.Open()
	if err != nil {
		ErrorLogger.Fatalln("Error while opening the session: ", err)
	}

	InfoLogger.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := session.ApplicationCommandCreate(session.State.User.ID, "", v)
		if err != nil {
			ErrorLogger.Panicf("Cannot create '%v' command: %v\n", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	defer session.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	InfoLogger.Println("Gracefully shutting down.")
}
