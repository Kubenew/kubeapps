import {
  AvailablePackageReference,
  Context,
  CreateInstalledPackageRequest,
  DeleteInstalledPackageRequest,
  InstalledPackageReference,
  ReconciliationOptions,
  UpdateInstalledPackageRequest,
  VersionReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import * as url from "shared/url";
import { axiosWithAuth } from "./AxiosInstance";
import { KubeappsGrpcClient } from "./KubeappsGrpcClient";

export const KUBEOPS_ROOT_URL = "api/kubeops/v1";
export class App {
  private static client = () => new KubeappsGrpcClient().getPackagesServiceClientImpl();

  public static async GetInstalledPackageSummaries(
    cluster: string,
    namespace?: string,
    page?: number,
    size?: number,
  ) {
    return await this.client().GetInstalledPackageSummaries({
      context: { cluster: cluster, namespace: namespace },
      paginationOptions: { pageSize: size || 0, pageToken: page?.toString() || "0" },
    });
  }

  public static async GetInstalledPackageDetail(installedPackageRef?: InstalledPackageReference) {
    return await this.client().GetInstalledPackageDetail({
      installedPackageRef: installedPackageRef,
    });
  }

  public static async createInstalledPackage(
    targetContext: Context,
    name: string,
    availablePackageRef: AvailablePackageReference,
    pkgVersionReference: VersionReference,
    values?: string,
    reconciliationOptions?: ReconciliationOptions,
  ) {
    return await this.client().CreateInstalledPackage({
      name,
      values,
      targetContext,
      availablePackageRef,
      pkgVersionReference,
      reconciliationOptions,
    } as CreateInstalledPackageRequest);
  }

  public static async updateInstalledPackage(
    installedPackageRef: InstalledPackageReference,
    pkgVersionReference: VersionReference,
    values?: string,
    reconciliationOptions?: ReconciliationOptions,
  ) {
    return await this.client().UpdateInstalledPackage({
      installedPackageRef,
      pkgVersionReference,
      values,
      reconciliationOptions,
    } as UpdateInstalledPackageRequest);
  }

  public static async rollback(installedPackageRef: InstalledPackageReference, revision: number) {
    const endpoint = url.kubeops.releases.get(
      installedPackageRef.context?.cluster ?? "",
      installedPackageRef.context?.namespace ?? "",
      installedPackageRef.identifier,
    );
    const { data } = await axiosWithAuth.put(
      endpoint,
      {},
      {
        params: {
          action: "rollback",
          revision,
        },
      },
    );
    return data;
  }

  public static async delete(installedPackageRef: InstalledPackageReference, purge: boolean) {
    let endpoint = url.kubeops.releases.get(
      installedPackageRef.context?.cluster ?? "",
      installedPackageRef.context?.namespace ?? "",
      installedPackageRef.identifier,
    );
    if (purge) {
      endpoint += "?purge=true";
    }
    const { data } = await axiosWithAuth.delete(endpoint);
    return data;
  }

  public static async deleteInstalledPackage(installedPackageRef: InstalledPackageReference) {
    return await this.client().DeleteInstalledPackage({
      installedPackageRef,
    } as DeleteInstalledPackageRequest);
  }

  // TODO(agamez): remove it once we return the generated resources as part of the InstalledPackageDetail.
  public static async getRelease(installedPackageRef?: InstalledPackageReference) {
    const { data } = await axiosWithAuth.get<{ data: { manifest: any } }>(
      url.kubeops.releases.get(
        installedPackageRef?.context?.cluster ?? "",
        installedPackageRef?.context?.namespace ?? "",
        installedPackageRef?.identifier ?? "",
      ),
    );
    return data.data;
  }
}
