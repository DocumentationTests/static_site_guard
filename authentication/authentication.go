package authentication

import (
	"github.com/g3ortega/hugo-auth/encrypt_tools"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/shareed2k/goth_fiber"
	"log"
	"os"
	"strings"
)

func Callback(ctx *fiber.Ctx, store *session.Store) error {
	user, err := goth_fiber.CompleteUserAuth(ctx)
	sess, err := store.Get(ctx)
	if err != nil {
		panic(err)
	}

	if err != nil {
		return ctx.Status(400).JSON(map[string]string{"message": err.Error()})
	} else {
		encryptedUsername := encrypt_tools.Encrypt(os.Getenv("SECRET_KEY"), user.NickName)
		sess.Set("userName", encryptedUsername)
		sess.Save()
	}

	if EligibleUser(user.NickName) {
		return ctx.Redirect("/")
	} else {
		return ctx.Redirect("/logout")
	}
}

func Logout(ctx *fiber.Ctx, store *session.Store) error {
	if err := goth_fiber.Logout(ctx); err != nil {
		log.Fatal(err)
	}

	sess, err := store.Get(ctx)
	if err != nil {
		panic(err)
	}

	userName := sess.Get("userName")
	log.Println(userName)

	if userName != nil {
		sess.Delete("userName")
		sess.Destroy()
		sess.Save()
	}

	return ctx.Redirect("/login")
}

func EligibleUser(userName string) bool {
	eligibleUsernames := os.Getenv("ELIGIBLE_USERNAMES")
	if eligibleUsernames == "" {
		log.Fatal("ELIGIBLE_USERNAMES is not set")
	}

	arr := strings.Split(eligibleUsernames, ",")

	if contains(arr, userName) {
		return true
	} else {
		return false
	}
}

func SessionHandler(ctx *fiber.Ctx, store *session.Store) error {
	sess, err := store.Get(ctx)
	if err != nil {
		panic(err)
	}
	encryptedUsername := sess.Get("userName")
	userName := ""

	if encryptedUsername != nil {
		userName = encrypt_tools.Decrypt(os.Getenv("SECRET_KEY"), encryptedUsername.(string))
	}

	if EligibleUser(userName) {
		return ctx.Next()
	} else {
		if ctx.Path() == "/login" || ctx.Path() == "/login/github" || ctx.Path() == "/not_authorized" || ctx.Path() == "/auth/callback/github" || ctx.Path() == "/logout" {
			return ctx.Next()
		} else {
			return ctx.Redirect("/not_authorized")
		}
	}
}

func contains(arr []string, value string) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}
