package middleware

import (
	"github.com/gofiber/fiber/v2"

	"api-students/helper"
)

func RequirePermission(perms *helper.PermissionSet, permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := helper.CurrentUser(c)
		if !ok {
			return helper.Unauthorized("belum terautentikasi")
		}

		if !perms.Can(user.Role, permission) {
			return helper.Forbidden("role " + user.Role + " tidak memiliki hak " + permission)
		}

		return c.Next()
	}
}
