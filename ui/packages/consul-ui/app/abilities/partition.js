import BaseAbility from 'consul-ui/abilities/base';
import { inject as service } from '@ember/service';

export default class PartitionAbility extends BaseAbility {
  @service('env') env;

  resource = 'operator';
  segmented = false;

  get isLinkable() {
    return !this.item.DeletedAt;
  }

  get canManage() {
    return this.canCreate;
  }

  get canDelete() {
    return this.item.Name !== 'default' && super.canDelete;
  }

  get canChoose() {
    return this.canUse;
  }

  get canUse() {
    return this.env.var('CONSUL_PARTITIONS_ENABLED');
  }
}