import { CdsIcon } from "@cds/react/icon";
import Alert from "components/js/Alert";
import { InstalledPackageDetail } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Link } from "react-router-dom";
import { app as appURL } from "shared/url";
interface IChartInfoProps {
  installedPackageDetail: InstalledPackageDetail;
}

export default function ChartUpdateInfo({ installedPackageDetail }: IChartInfoProps) {
  let alertContent;
  if (
    installedPackageDetail.latestVersion?.appVersion &&
    installedPackageDetail.currentVersion?.appVersion &&
    installedPackageDetail.currentVersion?.appVersion !==
      installedPackageDetail.latestVersion?.appVersion
  ) {
    // There is a new application version
    alertContent = (
      <>
        A new app version is available:{" "}
        <strong>{installedPackageDetail.latestVersion?.appVersion}</strong>.{" "}
      </>
    );
  } else if (
    installedPackageDetail.latestVersion?.pkgVersion &&
    installedPackageDetail.currentVersion?.pkgVersion &&
    installedPackageDetail.latestVersion?.pkgVersion !==
      installedPackageDetail.currentVersion?.pkgVersion
  ) {
    // There is a new package version
    alertContent = (
      <>
        A new package version is available:{" "}
        <strong>{installedPackageDetail.latestVersion?.pkgVersion}</strong>.{" "}
      </>
    );
  }
  // App is up to date
  return alertContent ? (
    <Alert>
      {alertContent}
      <Link to={appURL.apps.upgrade(installedPackageDetail.installedPackageRef)}>Update Now</Link>
    </Alert>
  ) : (
    <div className="color-icon-success">
      <CdsIcon shape="check-circle" size="md" solid={true} /> Up to date
    </div>
  );
}
