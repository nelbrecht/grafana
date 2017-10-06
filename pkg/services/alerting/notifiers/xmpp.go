package notifiers

import (
	"strings"
	"fmt"

	"github.com/grafana/grafana/pkg/log"
	"github.com/grafana/grafana/pkg/setting"
	m "github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/alerting"

	"github.com/mattn/go-xmpp"
)

func init() {
	alerting.RegisterNotifier(&alerting.NotifierPlugin{
		Type:        "xmpp",
		Name:        "XMPP (Jabber)",
		Description: "Sends xmpp messages",
		Factory:     NewXMPPNotifier,
		OptionsTemplate: `
      <h3 class="page-heading"> XMPP/Jabber settings</h3>
      <h4>Users</h4>
      <div class="gf-form">
        <textarea rows="7" class="gf-form-input width-25" ng-model="ctrl.model.settings.users"></textarea>
      </div>
      <h4>MUC</h4>
      <div class="gf-form">
        <textarea rows="7" class="gf-form-input width-25" ng-model="ctrl.model.settings.mucs"></textarea>
      </div>
    `,
	})

}

func NewXMPPNotifier(model *m.AlertNotification) (alerting.Notifier, error) {
	usersString := model.Settings.Get("users").MustString()
	users := strings.FieldsFunc(usersString, func(r rune) bool {
		switch r {
			case ',', ';', '\n':
				return true
		}
		return false
	})

	mucsString := model.Settings.Get("mucs").MustString()
	mucs := strings.FieldsFunc(mucsString, func(r rune) bool {
		switch r {
			case ',', ';', '\n':
				return true
		}
		return false
	})

	return &XMPPNotifier{
		NotifierBase: NewNotifierBase(model.Id, model.IsDefault, model.Name, model.Type, model.Settings),
		Users:        users,
		MUCs:         mucs,
		log:          log.New("alerting.notifier.xmpp"),
	}, nil
}

type XMPPNotifier struct {
	NotifierBase
	Users      []string
	MUCs       []string
	log        log.Logger
}

func (this *XMPPNotifier) Notify(evalContext *alerting.EvalContext) error {
	this.log.Info("Sending xmpp")

	if !setting.XMPPEnabled {
		return nil
	}

	client, err := setting.XMPP.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()

	msg := fmt.Sprintf("%s\n\n%s", evalContext.GetNotificationTitle(), evalContext.Rule.Message)

	for _, muc := range this.MUCs {
		client.JoinMUCNoHistory(muc,"")
		client.SendHtml(xmpp.Chat{Remote: muc, Type: "groupchat", Text: msg})
		client.LeaveMUC(muc)
	}
	for _, user := range this.Users {
		client.SendHtml(xmpp.Chat{Remote: user, Type: "chat", Text: msg})
	}

	if evalContext.ImagePublicUrl == "" {
		this.log.Info(fmt.Sprintf("Sending xmpp messages at %d mucs and at %d users", len(this.MUCs), len(this.Users)))
		return nil
	}

	msg = fmt.Sprintf(`<message to='%%s' type='%%s'>
		<body>%s</body>
		<x xmlns='jabber:x:oob'>
			<url>%s</url>
		</x>
	</message>`, evalContext.ImagePublicUrl, evalContext.ImagePublicUrl)

	for _, muc := range this.MUCs {
		client.JoinMUCNoHistory(muc,"")
		client.SendOrg(fmt.Sprintf(msg, muc, "groupchat"))
		client.LeaveMUC(muc)
	}
	for _, user := range this.Users {
		client.SendOrg(fmt.Sprintf(msg, user, "chat"))
	}
	this.log.Info(fmt.Sprintf("Sending xmpp messages and images at %d mucs and at %d users", len(this.MUCs), len(this.Users)))

	return nil
}
