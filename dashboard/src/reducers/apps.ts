import { LocationChangeAction, LOCATION_CHANGE } from "connected-react-router";
import { InstalledPackageDetailCustomDataHelm } from "gen/kubeappsapis/plugins/helm/packages/v1alpha1/helm";
import { IAppState } from "shared/types";
import { getType } from "typesafe-actions";
import actions from "../actions";
import { AppsAction } from "../actions/apps";

export const initialState: IAppState = {
  isFetching: false,
  items: [],
};

const appsReducer = (
  state: IAppState = initialState,
  action: AppsAction | LocationChangeAction,
): IAppState => {
  switch (action.type) {
    case getType(actions.apps.requestApps):
      return { ...state, isFetching: true };
    case getType(actions.apps.errorApp):
      return { ...state, isFetching: false, error: action.payload };
    case getType(actions.apps.selectApp):
      /* eslint-disable-next-line no-case-declarations */
      let revision: number;
      try {
        // TODO(agamez): verify why the field is not automatically decoded.
        revision = InstalledPackageDetailCustomDataHelm.decode(
          action.payload.app?.customDetail?.value as unknown as Uint8Array,
        ).releaseRevision;
      } catch (error) {
        // If the decoding fails, ignore it and just fall back to "no revisions"
        revision = 0;
      }
      return {
        ...state,
        isFetching: false,
        selected: {
          ...action.payload.app,
          // TODO(agamez): remove it once we have a core mechanism for rolling back
          revision: revision,
          // TODO(agamez): remove it once we return the generated resources as part of the InstalledPackageDetail.
          manifest: action.payload.manifest,
        },
        selectedDetails: action.payload.details,
      };
    case getType(actions.apps.listApps):
      return { ...state, isFetching: true };
    case getType(actions.apps.receiveAppList):
      return { ...state, isFetching: false, listOverview: action.payload };
    case getType(actions.apps.requestDeleteApp):
      return { ...state, isFetching: true };
    case getType(actions.apps.receiveDeleteApp):
      return { ...state, isFetching: false };
    case getType(actions.apps.requestDeployApp):
      return { ...state, isFetching: true };
    case getType(actions.apps.receiveDeployApp):
      return { ...state, isFetching: false };
    case getType(actions.apps.requestRollbackApp):
      return { ...state, isFetching: true };
    case getType(actions.apps.receiveRollbackApp):
      return { ...state, isFetching: false };
    case LOCATION_CHANGE:
      return {
        ...state,
        error: undefined,
        isFetching: false,
        selected: undefined,
      };
    default:
  }
  return state;
};

export default appsReducer;
