package middleware

import (
	"context"
	"fmt"
	"net/http"

	"gantt/internal/interactor/pkg/util/log"

	"github.com/gin-gonic/gin"
	"github.com/open-policy-agent/opa/rego"
)

func WithOPA(opa *rego.PreparedEvalQuery) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := c.Query("user")
		groups := c.QueryArray("groups")
		input := map[string]interface{}{ // 构造OPA输入
			"method": c.Request.Method,
			"path":   c.Request.RequestURI,
			"subject": map[string]interface{}{
				"user":  user,
				"group": groups,
			},
		}

		log.Info(fmt.Sprintf("start opa middleware %s, %#v", c.Request.URL.String(), input))
		res, err := opa.Eval(context.TODO(), rego.EvalInput(input))
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			c.Abort()
			return
		}

		defer log.Info(fmt.Sprintf("opa result: %v, %#v", res.Allowed(), res))
		if !res.Allowed() {
			c.JSON(http.StatusForbidden, gin.H{
				"msg": "forbidden",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
