import React, { useEffect, useState } from "react";
import { List, ListItem, ListItemText } from "@mui/material";
import axios from "axios";
import type { Deployment } from "./types";
import { API_BASE_URL } from "./consts";

const DeploymentList: React.FC = () => {
  const [deployments, setDeployments] = useState<Deployment[]>([]);

  useEffect(() => {
    axios.get(API_BASE_URL + "/deployments").then((response) => {
      setDeployments(response.data);
    });
  }, []);

  return (
    <List>
      {deployments.map((dep) => (
        <ListItem key={dep.id}>
          <ListItemText
            primary={`${dep.name} (${dep.mode})`}
            secondary={`Ingress: ${dep.ingress} | Container: ${dep.container_id.substring(0, 12)}`}
          />
        </ListItem>
      ))}
    </List>
  );
};

export default DeploymentList;
