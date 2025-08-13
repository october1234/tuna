export interface Deployment {
  id: number;
  name: string;
  mode: string;
  repo?: string;
  branch?: string;
  image?: string;
  container_id: string;
  env: { [key: string]: string };
  ingress: string;
}

export interface DeployRequest {
  name: string;
  mode: string;
  repo?: string;
  branch?: string;
  template?: string;
  image?: string;
  env: { [key: string]: string };
  ingress: string;
}
