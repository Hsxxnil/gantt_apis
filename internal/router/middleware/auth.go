package middleware

import (
	policyModel "gantt/internal/interactor/models/policies"
	"gantt/internal/interactor/pkg/connect"
	"net/http"

	"github.com/casbin/casbin/v2"

	_ "gantt/internal/interactor/pkg/connect"
	"gantt/internal/interactor/pkg/util/log"

	"github.com/casbin/casbin/v2/model"

	gormAdapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gin-gonic/gin"
	_ "gorm.io/driver/postgres"
)

func newAdapter() *gormAdapter.Adapter {
	db, err := connect.PostgresSQL()
	if err != nil {
		log.Error(err)
		panic(err)
	}

	adapter, err := gormAdapter.NewAdapterByDB(db)
	if err != nil {
		log.Error(err)
		panic(err)
	}

	return adapter
}

func newEnforcer(adapter *gormAdapter.Adapter) *casbin.Enforcer {
	cmodel, err := model.NewModelFromString(`[request_definition]
	r = sub, obj, act

	[policy_definition]
	p = sub, obj, act

	[policy_effect]
	e = some(where (p.eft == allow))

	[matchers]
	m = r.sub == p.sub && keyMatch(r.obj, p.obj) && regexMatch(r.act, p.act)

	#[matchers]
	#m = r.sub == p.sub && r.obj == p.obj && r.act == p.act`)
	if err != nil {
		log.Error(err)
		panic(err)
	}

	enforcer, err := casbin.NewEnforcer(cmodel, adapter)
	if err != nil {
		log.Error(err)
		panic(err)
	}

	return enforcer
}

var Enforcer *casbin.Enforcer

func init() {
	adapter := newAdapter()
	Enforcer = newEnforcer(adapter)
}

func CreatePolicy(pm []*policyModel.PolicyModel) (bool, error) {
	var policies [][]string
	for _, p := range pm {
		policies = append(policies, []string{p.RoleName, p.Path, p.Method})
	}

	result, err := Enforcer.AddPolicies(policies)
	if err != nil {
		return false, err
	}

	err = Enforcer.LoadPolicy()
	if err != nil {
		return false, err
	}

	return result, nil
}

func DeletePolicy(pm []*policyModel.PolicyModel) (bool, error) {
	var policies [][]string
	for _, p := range pm {
		policies = append(policies, []string{p.RoleName, p.Path, p.Method})
	}

	result, err := Enforcer.RemovePolicies(policies)
	if err != nil {
		return false, err
	}

	err = Enforcer.LoadPolicy()
	if err != nil {
		return false, err
	}

	return result, err
}

func GetAllPolicies() ([][]string, error) {
	return Enforcer.GetPolicy()
}

func CheckPermission() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		log.Info("Casbin policies:", ctx.MustGet("role").(string), ctx.Request.URL.Path, ctx.Request.Method)
		res, err := Enforcer.Enforce(ctx.MustGet("role").(string), ctx.Request.URL.Path, ctx.Request.Method)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status": -1,
				"msg":    err.Error(),
			})
			ctx.Abort()
			return
		}

		if res {
			ctx.Next()
		} else {
			ctx.JSON(http.StatusNonAuthoritativeInfo, gin.H{
				"status": 203,
				"msg":    "Sorry, you don't have permission.",
			})
			ctx.Abort()
			return
		}
	}
}
