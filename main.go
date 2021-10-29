package main

import (
	"context"
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/MemeLabs/dggchat"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	debuglogger = log.New(os.Stdout, "[d] ", log.Ldate|log.Ltime|log.Lshortfile)
	authCookie  string
	chatURL     string
	mongoURL    string
)

func main() {
	flag.StringVar(&authCookie, "cookie", "", "Cookie used for chat authentication and API access")
	flag.StringVar(&chatURL, "chat", "wss://chat.destiny.gg/ws", "ws(s)-url for chat")
	flag.StringVar(&mongoURL, "mongo", "", "mongo token")
	flag.Parse()

	mongoOptions := options.Client().ApplyURI(mongoURL)
	mongoClient, err := mongo.NewClient(mongoOptions)
	if err != nil {
		log.Fatalln(err)
	}

	dgg, err := dggchat.New(";jwt=" + authCookie)
	if err != nil {
		log.Fatalln(err)
	}

	// init bot
	b := newBot(authCookie, 250)

	b.mongo = *mongoClient
	b.mongo.Connect(context.TODO())

	users, err := b.getOptedUsers()
	if err != nil {
		log.Fatalln(err)
	}
	b.optedUsers = users

	defer b.mongo.Disconnect(context.TODO())

	b.addMSGParser(
		b.say,
		b.isUserMention,
		b.getMentions,
		b.optIn,
	)
	b.addPRIVMSGParser()
	dgg.AddMessageHandler(b.onMessage)
	dgg.AddPMHandler(b.onPMHandler)
	dgg.AddErrorHandler(b.onError)
	dgg.AddMuteHandler(b.onMute)
	dgg.AddUnmuteHandler(b.onUnmute)
	dgg.AddBanHandler(b.onBan)
	dgg.AddUnbanHandler(b.onUnban)
	dgg.AddSocketErrorHandler(b.onSocketError)

	u, err := url.Parse(chatURL)
	if err != nil {
		log.Fatalln(err)
	}
	dgg.SetURL(*u)

	err = dgg.Open()
	if err != nil {
		log.Fatalln(err)
	}
	debuglogger.Println("[##] connected...")
	defer dgg.Close()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	debuglogger.Println("[##] waiting for signals...")
	for {
		sig := <-signals
		switch sig {

		// exit on interrupt
		case syscall.SIGTERM:
			fallthrough
		case syscall.SIGINT:
			log.Println("[##] signal: handling SIGINT/SIGTERM")
			os.Exit(1)
		}
	}
}
