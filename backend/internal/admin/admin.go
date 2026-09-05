package admin

import "github.com/wanglei-123-wl/ika6server/backend/internal/users"

func CanManage(user users.User) bool {
	return user.Role == users.RoleAdmin
}
