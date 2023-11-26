package auth

import (
	policyModel "hta/internal/interactor/models/policies"
	"hta/internal/interactor/pkg/connect"
	"net/http"

	"github.com/casbin/casbin/v2"

	"github.com/casbin/casbin/v2/model"
	_ "hta/internal/interactor/pkg/connect"
	"hta/internal/interactor/pkg/util/log"

	gormAdapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gin-gonic/gin"
	_ "gorm.io/driver/postgres"
)

func newAdapter() *gormAdapter.Adapter {
	db, err := connect.PostgresSQL()
	if err != nil {
		panic(err)
	}

	adapter, err := gormAdapter.NewAdapterByDB(db)
	if err != nil {
		panic(err)
	}

	//cb := []byte(_CASBIN_RULES)
	//
	//a := jsonAdapter.NewAdapter(&cb)

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
		log.Info("model error: %s;", err)
	}

	enforcer, err := casbin.NewEnforcer(cmodel, adapter)
	if err != nil {
		panic(err)
	}

	return enforcer
}

var Enforcer *casbin.Enforcer

func init() {
	adapter := newAdapter()
	Enforcer = newEnforcer(adapter)
}

func CreatePolicy(pm policyModel.PolicyModel) (bool, error) {
	result, err := Enforcer.AddPolicy(pm.RoleName, pm.Path, pm.Method)
	// 重新加載policy
	err = Enforcer.LoadPolicy()
	if err != nil {
		return false, err
	}

	return result, err
}

func DeletePolicy(pm policyModel.PolicyModel) (bool, error) {
	result, err := Enforcer.RemovePolicy(pm.RoleName, pm.Path, pm.Method)
	// 重新加載policy
	err = Enforcer.LoadPolicy()
	if err != nil {
		return false, err
	}

	return result, err
}

func GetAllPolicies() [][]string {
	return Enforcer.GetPolicy()
}

func CheckPermission() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 去除path中的uuid
		//array := strings.Split(ctx.Request.URL.Path, "/")
		//var path string
		//if _, err := uuid.Parse(array[len(array)-1]); err == nil {
		//	path = strings.Join(array[:len(array)-1], "/")
		//} else {
		//	path = ctx.Request.URL.Path
		//}

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
