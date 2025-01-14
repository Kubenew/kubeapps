import { CdsProgressCircle } from "@cds/react/progress-circle";
import "./LoadingWrapper.css";

export interface ILoadingWrapperProps {
  loaded?: boolean;
  size?: string;
  loadingText?: string | JSX.Element;
  children?: any;
  className?: string;
}

function LoadingWrapper(props: ILoadingWrapperProps) {
  return props.loaded ? (
    props.children
  ) : (
    <div className={props.className || ""}>
      {props.loadingText && <div className="flex-h-center loadingText">{props.loadingText}</div>}
      <div className="flex-h-center margin-t-md">
        <CdsProgressCircle size={props.size || "xxl"} status="info" />
      </div>
    </div>
  );
}

export default LoadingWrapper;
