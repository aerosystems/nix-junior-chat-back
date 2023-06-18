package FirebaseClient

import (
	"context"
	"log"
	"os"
	"path/filepath"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"google.golang.org/api/option"
)

func SetupFirebase() (*firebase.App, context.Context, *messaging.Client) {

	ctx := context.Background()

	serviceAccountKeyFilePath, err := filepath.Abs(os.Getenv("FIREBASE_KEY_PATH"))
	if err != nil {
		log.Println("Unable to load serviceAccountKeys.json file")
		os.Exit(1)
	}

	opt := option.WithCredentialsFile(serviceAccountKeyFilePath)

	//Firebase admin SDK initialization
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	//Messaging client
	client, _ := app.Messaging(ctx)

	return app, ctx, client
}
