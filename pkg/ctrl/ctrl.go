// SPDX-License-Identifier: AGPL-3.0-or-later
package ctrl

import (
	"github.com/lanthora/uranus/pkg/connector"
)

func Shutdown() bool {
	connector.Exec(`{"type":"user::ctrl::exit"}`, 0)
	return true
}
