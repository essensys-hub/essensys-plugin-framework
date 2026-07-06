package plugin

import "fmt"

// TopicBase est la racine des topics de plugins sur le bus.
const TopicBase = "essensys/plugins"

// TopicFilter renvoie le filtre d'abonnement d'un plugin (toutes armoires).
//
//	essensys/plugins/<id>/+/#
func TopicFilter(pluginID string) string {
	return fmt.Sprintf("%s/%s/+/#", TopicBase, pluginID)
}

// Topic construit le topic d'une mesure pour une armoire.
//
//	essensys/plugins/<id>/<machine_id>/<metric>
func Topic(pluginID, machineID, metric string) string {
	return fmt.Sprintf("%s/%s/%s/%s", TopicBase, pluginID, machineID, metric)
}

// HeartbeatTopic est le topic de heartbeat d'un collecteur.
func HeartbeatTopic(pluginID, machineID string) string {
	return fmt.Sprintf("%s/%s/%s/_heartbeat", TopicBase, pluginID, machineID)
}
