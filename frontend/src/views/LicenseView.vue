<template>
  <div class="license-container">
    <h1>License管理系统</h1>
    
    <div class="license-actions">
      <el-button type="primary" @click="showAddDialog = true">新增License</el-button>
      <el-button @click="refreshData">刷新数据</el-button>
    </div>

    <!-- License统计图表 -->
    <div class="license-chart">
      <div id="licenseChart" style="width: 100%; height: 300px;"></div>
    </div>

    <!-- License列表 -->
    <div class="license-table">
      <el-table :data="licenseList" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="customer" label="客户名称" width="150" />
        <el-table-column prop="fingerprint" label="硬件指纹" width="180" />
        <el-table-column prop="activated_at" label="激活时间" width="180" />
        <el-table-column prop="expires_at" label="过期时间" width="180" />
        <el-table-column prop="is_active" label="状态" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.is_active ? 'success' : 'danger'">
              {{ scope.row.is_active ? '激活' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150">
          <template #default="scope">
            <el-button size="small" @click="deactivateLicense(scope.row.id)" v-if="scope.row.is_active">
              停用
            </el-button>
            <el-button size="small" type="danger" @click="deleteLicense(scope.row.id)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 新增License对话框 -->
    <el-dialog v-model="showAddDialog" title="新增License" width="500px">
      <el-form :model="newLicense" label-width="120px">
        <el-form-item label="客户名称">
          <el-input v-model="newLicense.customer" placeholder="请输入客户名称" />
        </el-form-item>
        <el-form-item label="硬件指纹">
          <el-input v-model="newLicense.fingerprint" placeholder="请输入硬件指纹" />
        </el-form-item>
        <el-form-item label="有效期(天)">
          <el-input-number v-model="newLicense.validityDays" :min="1" :max="3650" />
        </el-form-item>
        <el-form-item label="License内容">
          <el-input v-model="newLicense.licenseContent" type="textarea" rows="4" placeholder="请输入License内容" />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showAddDialog = false">取消</el-button>
          <el-button type="primary" @click="addLicense">确定</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import * as echarts from 'echarts'
import { ElMessage, ElMessageBox } from 'element-plus'
import axios from 'axios'

// 数据定义
const licenseList = ref([])
const showAddDialog = ref(false)
const chartInstance = ref(null)

const newLicense = ref({
  customer: '',
  fingerprint: '',
  validityDays: 365,
  licenseContent: ''
})

// API基础URL
const API_BASE_URL = 'http://localhost:8080/api'

// 初始化图表
const initChart = () => {
  const chartDom = document.getElementById('licenseChart')
  if (chartDom) {
    chartInstance.value = echarts.init(chartDom)
    updateChart()
  }
}

// 更新图表数据
const updateChart = () => {
  if (!chartInstance.value) return
  
  // 计算统计数据
  const totalCount = licenseList.value.length
  const activeCount = licenseList.value.filter(item => item.is_active).length
  const expiredCount = totalCount - activeCount
  
  const option = {
    title: {
      text: 'License统计',
      left: 'center'
    },
    tooltip: {
      trigger: 'item'
    },
    legend: {
      orient: 'vertical',
      left: 'left'
    },
    series: [
      {
        name: 'License状态',
        type: 'pie',
        radius: '50%',
        data: [
          { value: activeCount, name: '激活中' },
          { value: expiredCount, name: '已过期/停用' }
        ],
        emphasis: {
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: 'rgba(0, 0, 0, 0.5)'
          }
        }
      }
    ]
  }
  
  chartInstance.value.setOption(option)
}

// 获取License列表
const fetchLicenseList = async () => {
  try {
    const response = await axios.get(`${API_BASE_URL}/license/activations`)
    // 后端返回格式是 {"activations": [...] }
    licenseList.value = response.data.activations || []
    updateChart()
  } catch (error) {
    console.error('获取License列表失败:', error)
    ElMessage.error('获取License列表失败')
  }
}

// 添加License
const addLicense = async () => {
  try {
    // 首先激活License
    const activateResponse = await axios.post(`${API_BASE_URL}/license/activate`, {
      license: newLicense.value.licenseContent
    })
    
    if (activateResponse.data.success) {
      ElMessage.success('License添加成功')
      showAddDialog.value = false
      // 重置表单
      newLicense.value = {
        customer: '',
        fingerprint: '',
        validityDays: 365,
        licenseContent: ''
      }
      // 刷新列表
      fetchLicenseList()
    } else {
      ElMessage.error(activateResponse.data.message || 'License添加失败')
    }
  } catch (error) {
    console.error('添加License失败:', error)
    ElMessage.error('添加License失败')
  }
}

// 停用License
const deactivateLicense = async (id) => {
  try {
    await ElMessageBox.confirm('确定要停用这个License吗?', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    // 这里需要根据后端API实现停用功能
    // 目前后端可能没有提供停用特定License的API
    ElMessage.info('停用功能需要后端API支持')
  } catch (error) {
    // 用户取消操作
  }
}

// 删除License
const deleteLicense = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除这个License吗?', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    // 这里需要根据后端API实现删除功能
    // 目前后端可能没有提供删除特定License的API
    ElMessage.info('删除功能需要后端API支持')
  } catch (error) {
    // 用户取消操作
  }
}

// 刷新数据
const refreshData = () => {
  fetchLicenseList()
  ElMessage.success('数据已刷新')
}

// 组件挂载时初始化
onMounted(() => {
  fetchLicenseList()
  initChart()
  
  // 窗口大小变化时重新渲染图表
  window.addEventListener('resize', () => {
    if (chartInstance.value) {
      chartInstance.value.resize()
    }
  })
})
</script>

<style scoped>
.license-container {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

.license-actions {
  margin-bottom: 20px;
  display: flex;
  gap: 10px;
}

.license-chart {
  margin-bottom: 30px;
  background: #fff;
  border-radius: 4px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  padding: 20px;
}

.license-table {
  background: #fff;
  border-radius: 4px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  padding: 20px;
}

h1 {
  color: #333;
  margin-bottom: 20px;
  text-align: center;
}
</style>